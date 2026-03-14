/**
 * tools/getTaskByFilter.js
 * Tool to get tasks/subtasks by flexible filters across projects
 */

import { z } from "zod";
import {
  executeTaskMasterCommand,
  createContentResponse,
  createErrorResponse,
} from "./utils.js";

/**
 * Register the getTaskByFilter tool with the MCP server
 * @param {Object} server - FastMCP server instance
 */
export function registerGetTaskByFilterTool(server) {
  server.addTool({
    name: "getTaskByFilter",
    description:
      "Get tasks or subtasks from any project using flexible filters. Can filter by task ID, status, priority, title keywords, or get specific subtasks from another project.",
    parameters: z.object({
      projectRoot: z
        .string()
        .describe("Root directory of the project to read tasks from"),
      taskId: z
        .string()
        .optional()
        .describe(
          "Specific task ID to retrieve (e.g., '1' for task 1, '1.2' for subtask 2 of task 1)"
        ),
      status: z
        .string()
        .optional()
        .describe("Filter by status (pending, done, deferred, in-progress)"),
      priority: z
        .string()
        .optional()
        .describe("Filter by priority (high, medium, low)"),
      keyword: z
        .string()
        .optional()
        .describe(
          "Filter tasks by keyword in title or description (supports both English and Chinese)"
        ),
      includeSubtasks: z
        .boolean()
        .optional()
        .describe("Whether to include subtasks in the result"),
      subtaskOnly: z
        .boolean()
        .optional()
        .describe("Only return subtasks (not parent tasks)"),
      file: z.string().optional().describe("Path to the tasks file"),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(`Getting tasks with filters: ${JSON.stringify(args)}`);

        const cmdArgs = [];

        // If specific task ID is provided, use show command
        if (args.taskId) {
          const result = executeTaskMasterCommand(
            "show",
            log,
            [args.taskId],
            args.projectRoot
          );

          if (!result.success) {
            throw new Error(result.error);
          }

          return createContentResponse(result.stdout);
        }

        // Otherwise, use list command with filters
        if (args.status) cmdArgs.push(`--status=${args.status}`);
        if (args.includeSubtasks) cmdArgs.push("--with-subtasks");
        if (args.file) cmdArgs.push(`--file=${args.file}`);

        const result = executeTaskMasterCommand(
          "list",
          log,
          cmdArgs,
          args.projectRoot
        );

        if (!result.success) {
          throw new Error(result.error);
        }

        // Parse the output and apply additional filters
        let tasks = result.stdout;

        // Apply keyword filter if provided
        if (args.keyword) {
          const keyword = args.keyword.toLowerCase();
          const lines = tasks.split("\n");
          const filteredLines = lines.filter((line) => {
            const lineLower = line.toLowerCase();
            return lineLower.includes(keyword);
          });
          tasks = filteredLines.join("\n");
        }

        // Apply priority filter if provided
        if (args.priority) {
          const lines = tasks.split("\n");
          const filteredLines = lines.filter((line) => {
            return line.toLowerCase().includes(`priority: ${args.priority}`);
          });
          tasks = filteredLines.join("\n");
        }

        // Apply subtask-only filter if provided
        if (args.subtaskOnly) {
          const lines = tasks.split("\n");
          const subtaskLines = lines.filter((line) => {
            // Subtasks typically have IDs like "1.1", "1.2", etc.
            return /^\s*\d+\.\d+/.test(line);
          });
          tasks = subtaskLines.join("\n");
        }

        if (!tasks.trim()) {
          tasks = "No tasks found matching the specified filters.";
        }

        return createContentResponse(tasks);
      } catch (error) {
        log.error(`Error getting tasks: ${error.message}`);
        return createErrorResponse(`Error getting tasks: ${error.message}`);
      }
    },
  });
}
