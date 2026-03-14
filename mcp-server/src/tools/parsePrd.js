/**
 * tools/parsePrd.js
 * Tool to parse PRD (Product Requirements Document) and generate tasks
 */

import { z } from "zod";
import {
  executeTaskMasterCommand,
  createContentResponse,
  createErrorResponse,
} from "./utils.js";

/**
 * Register the parsePrd tool with the MCP server
 * @param {Object} server - FastMCP server instance
 */
export function registerParsePrdTool(server) {
  server.addTool({
    name: "parsePrd",
    description:
      "Parse a PRD (Product Requirements Document) file and generate tasks with AI. This tool can read requirements from any project directory and generate structured tasks.",
    parameters: z.object({
      prdFile: z
        .string()
        .describe(
          "Path to the PRD file (can be absolute path or relative to project root)"
        ),
      numTasks: z
        .number()
        .optional()
        .describe(
          "Number of tasks to generate (optional - AI will determine based on complexity if not specified)"
        ),
      knowledgeBase: z
        .string()
        .optional()
        .describe(
          "Path to business knowledge documents (file or directory) to use as context"
        ),
      outputFile: z
        .string()
        .optional()
        .describe("Output file path for tasks.json (default: tasks/tasks.json)"),
      projectRoot: z
        .string()
        .describe(
          "Root directory of the project where tasks should be generated (default: current working directory)"
        ),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(`Parsing PRD file: ${args.prdFile}`);

        const cmdArgs = [];

        // Add PRD file path
        cmdArgs.push(args.prdFile);

        // Add optional parameters
        if (args.numTasks) {
          cmdArgs.push(`--num-tasks=${args.numTasks}`);
        }

        if (args.knowledgeBase) {
          cmdArgs.push(`--knowledge-base=${args.knowledgeBase}`);
        }

        if (args.outputFile) {
          cmdArgs.push(`--output=${args.outputFile}`);
        }

        const projectRoot = args.projectRoot;

        const result = executeTaskMasterCommand(
          "parse-prd",
          log,
          cmdArgs,
          projectRoot
        );

        if (!result.success) {
          throw new Error(result.error);
        }

        log.info(`PRD parsing completed successfully`);

        return createContentResponse(
          `Successfully parsed PRD and generated tasks.\n\n${result.stdout}`
        );
      } catch (error) {
        log.error(`Error parsing PRD: ${error.message}`);
        return createErrorResponse(`Error parsing PRD: ${error.message}`);
      }
    },
  });
}
