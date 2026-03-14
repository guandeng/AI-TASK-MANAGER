/**
 * tools/getStructuredTasks.js
 * Tool to get structured tasks optimized for AI code generation
 * Provides a clean, machine-readable format for another project to consume
 */

import { z } from "zod";
import {
  createContentResponse,
  createErrorResponse,
} from "./utils.js";
import fs from "fs";
import path from "path";

/**
 * Get task data from storage (file or database)
 * @param {string} projectRoot - Project root directory
 * @returns {Object|null} - Task data or null
 */
function getTaskData(projectRoot) {
  // Try file storage first
  const tasksPath = path.join(projectRoot, "tasks", "tasks.json");
  if (fs.existsSync(tasksPath)) {
    return JSON.parse(fs.readFileSync(tasksPath, "utf8"));
  }

  // Try task.json in root (alternative location)
  const altTasksPath = path.join(projectRoot, "task.json");
  if (fs.existsSync(altTasksPath)) {
    return JSON.parse(fs.readFileSync(altTasksPath, "utf8"));
  }

  return null;
}

/**
 * Transform task to AI-friendly format
 * @param {Object} task - Original task object
 * @param {number} index - Task index
 * @returns {Object} - Transformed task
 */
function transformTaskForAI(task, index) {
  const transformed = {
    // 唯一标识
    id: task.id || index + 1,
    fullId: `task-${task.id || index + 1}`,

    // 基本信息
    title: task.title || task.titleTrans || "",
    description: task.description || task.descriptionTrans || "",

    // 状态信息
    status: task.status || "pending",
    priority: task.priority || "medium",

    // 实现细节
    details: task.details || task.detailsTrans || "",
    testStrategy: task.testStrategy || task.testStrategyTrans || "",

    // 依赖关系
    dependencies: task.dependencies || [],

    // 元数据
    assignee: task.assignee || null,
    requirementId: task.requirementId || null,

    // 新增：代码生成相关字段
    techStack: task.techStack || [],
    relatedFiles: task.relatedFiles || [],
    codeHints: task.codeHints || "",
    executionOrder: task.executionOrder || null,
  };

  // 处理子任务
  if (task.subtasks && task.subtasks.length > 0) {
    transformed.subtasks = task.subtasks.map((subtask, subIndex) => ({
      id: subtask.id || subIndex + 1,
      fullId: `${transformed.id}.${subtask.id || subIndex + 1}`,
      title: subtask.title || subtask.titleTrans || "",
      description: subtask.description || subtask.descriptionTrans || "",
      details: subtask.details || subtask.detailsTrans || "",
      status: subtask.status || "pending",
      dependencies: subtask.dependencies || [],
      // 新增：代码生成相关字段
      relatedFiles: subtask.relatedFiles || [],
      codeInterface: subtask.codeInterface || null,
      acceptanceCriteria: subtask.acceptanceCriteria || [],
      codeHints: subtask.codeHints || "",
      executionOrder: subtask.executionOrder || null,
    }));
    transformed.subtaskCount = transformed.subtasks.length;
    transformed.completedSubtasks = transformed.subtasks.filter(
      (s) => s.status === "done"
    ).length;
  } else {
    transformed.subtasks = [];
    transformed.subtaskCount = 0;
    transformed.completedSubtasks = 0;
  }

  return transformed;
}

/**
 * Generate AI context prompt
 * @param {Object} task - Task object
 * @returns {string} - Context prompt
 */
function generateAIContext(task) {
  const lines = [];

  lines.push(`## Task: ${task.title}`);
  lines.push(`- ID: ${task.fullId}`);
  lines.push(`- Status: ${task.status}`);
  lines.push(`- Priority: ${task.priority}`);

  if (task.description) {
    lines.push(`\n### Description\n${task.description}`);
  }

  if (task.details) {
    lines.push(`\n### Implementation Details\n${task.details}`);
  }

  if (task.testStrategy) {
    lines.push(`\n### Test Strategy\n${task.testStrategy}`);
  }

  if (task.dependencies.length > 0) {
    lines.push(`\n### Dependencies\n${task.dependencies.join(", ")}`);
  }

  if (task.subtasks.length > 0) {
    lines.push(`\n### Subtasks (${task.completedSubtasks}/${task.subtaskCount} completed)`);
    task.subtasks.forEach((subtask) => {
      const statusIcon = subtask.status === "done" ? "✅" : subtask.status === "in-progress" ? "🔄" : "⏱️";
      lines.push(`- ${statusIcon} ${subtask.fullId}: ${subtask.title}`);
    });
  }

  return lines.join("\n");
}

/**
 * Register the getStructuredTasks tool with the MCP server
 * @param {Object} server - FastMCP server instance
 */
export function registerGetStructuredTasksTool(server) {
  server.addTool({
    name: "getStructuredTasks",
    description:
      "Get structured tasks optimized for AI code generation. Returns tasks in a clean, machine-readable format with all necessary context for another AI to understand and generate code. Perfect for cross-project task synchronization.",
    parameters: z.object({
      projectRoot: z
        .string()
        .describe("Root directory of the AI-TASK-MANAGER project"),
      requirementId: z
        .string()
        .optional()
        .describe("Filter by requirement ID (optional)"),
      taskId: z
        .string()
        .optional()
        .describe("Get specific task by ID (e.g., '1' or '1.2')"),
      status: z
        .string()
        .optional()
        .describe("Filter by status (pending, in-progress, done, deferred)"),
      priority: z
        .string()
        .optional()
        .describe("Filter by priority (high, medium, low)"),
      includeCompleted: z
        .boolean()
        .optional()
        .default(false)
        .describe("Include completed tasks (default: false)"),
      format: z
        .enum(["json", "prompt", "markdown"])
        .optional()
        .default("json")
        .describe("Output format: json (machine-readable), prompt (AI-optimized), markdown (human-readable)"),
      includeSubtasks: z
        .boolean()
        .optional()
        .default(true)
        .describe("Include subtasks in response"),
      includeContext: z
        .boolean()
        .optional()
        .default(true)
        .describe("Include AI context for code generation"),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(`Getting structured tasks from: ${args.projectRoot}`);

        const data = getTaskData(args.projectRoot);
        if (!data) {
          return createErrorResponse(
            `No task data found in ${args.projectRoot}. Expected tasks/tasks.json or task.json`
          );
        }

        const tasks = data.tasks || [];

        // Transform tasks
        let transformedTasks = tasks.map((task, index) =>
          transformTaskForAI(task, index)
        );

        // Apply filters
        if (args.requirementId) {
          transformedTasks = transformedTasks.filter(
            (t) => String(t.requirementId) === args.requirementId
          );
        }

        if (args.taskId) {
          const [mainId, subId] = args.taskId.toString().split(".");
          transformedTasks = transformedTasks.filter(
            (t) => String(t.id) === mainId
          );

          if (subId && transformedTasks.length > 0) {
            const task = transformedTasks[0];
            const subtask = task.subtasks.find((s) => String(s.id) === subId);
            if (subtask) {
              // Return single subtask
              const result = {
                type: "subtask",
                parentTask: {
                  id: task.id,
                  title: task.title,
                  status: task.status,
                },
                subtask: subtask,
              };
              return createContentResponse(JSON.stringify(result, null, 2));
            }
          }
        }

        if (args.status) {
          transformedTasks = transformedTasks.filter(
            (t) => t.status === args.status
          );
        }

        if (args.priority) {
          transformedTasks = transformedTasks.filter(
            (t) => t.priority === args.priority
          );
        }

        if (!args.includeCompleted) {
          transformedTasks = transformedTasks.filter(
            (t) => t.status !== "done"
          );
        }

        if (!args.includeSubtasks) {
          transformedTasks = transformedTasks.map((t) => {
            const { subtasks, ...rest } = t;
            return { ...rest, subtaskCount: t.subtaskCount, hasSubtasks: t.subtaskCount > 0 };
          });
        }

        // Build response based on format
        let response;

        if (args.format === "json") {
          // Pure JSON format - best for programmatic use
          response = {
            version: "1.0",
            source: {
              project: data.projectName || "AI Task Manager",
              version: data.projectVersion || "1.0.0",
              exportedAt: new Date().toISOString(),
            },
            summary: {
              totalTasks: transformedTasks.length,
              pending: transformedTasks.filter((t) => t.status === "pending").length,
              inProgress: transformedTasks.filter((t) => t.status === "in-progress").length,
              done: transformedTasks.filter((t) => t.status === "done").length,
            },
            tasks: transformedTasks,
          };

          if (args.includeContext) {
            response.aiContext = {
              usage:
                "These tasks are structured for AI code generation. Each task contains title, description, details, and testStrategy fields that provide context for implementation.",
              fields: {
                id: "Unique task identifier",
                fullId: "Full task identifier in format 'task-N' or 'N.M' for subtasks",
                title: "Short task title",
                description: "Detailed task description",
                details: "Implementation details and technical specifications",
                testStrategy: "How to test the implementation",
                dependencies: "IDs of tasks that must be completed first",
                subtasks: "Child tasks that break down this task",
              },
            };
          }

          return createContentResponse(JSON.stringify(response, null, 2));
        }

        if (args.format === "prompt") {
          // AI-optimized prompt format
          const lines = [];

          lines.push(`# Task Context for AI Code Generation`);
          lines.push(`Source: ${data.projectName || "AI Task Manager"}`);
          lines.push(`Exported: ${new Date().toISOString()}`);
          lines.push("");
          lines.push(`## Summary`);
          lines.push(
            `- Total Tasks: ${transformedTasks.length}`
          );
          lines.push(
            `- Pending: ${transformedTasks.filter((t) => t.status === "pending").length}`
          );
          lines.push(
            `- In Progress: ${transformedTasks.filter((t) => t.status === "in-progress").length}`
          );
          lines.push("");
          lines.push("---");
          lines.push("");

          transformedTasks.forEach((task) => {
            lines.push(generateAIContext(task));
            lines.push("");
            lines.push("---");
            lines.push("");
          });

          lines.push("## Instructions for AI");
          lines.push(
            "Use the above task information to understand the requirements and generate code accordingly."
          );
          lines.push(
            "Each task includes implementation details and test strategy to guide development."
          );
          lines.push(
            "Respect task dependencies - complete prerequisite tasks first."
          );

          return createContentResponse(lines.join("\n"));
        }

        // Markdown format (human-readable)
        const lines = [];
        lines.push(`# Structured Tasks Export`);
        lines.push("");
        lines.push(`**Project:** ${data.projectName || "AI Task Manager"}`);
        lines.push(`**Exported:** ${new Date().toISOString()}`);
        lines.push("");

        // Summary table
        lines.push(`## Summary`);
        lines.push("");
        lines.push("| Status | Count |");
        lines.push("|--------|-------|");
        lines.push(
          `| Pending | ${transformedTasks.filter((t) => t.status === "pending").length} |`
        );
        lines.push(
          `| In Progress | ${transformedTasks.filter((t) => t.status === "in-progress").length} |`
        );
        lines.push(
          `| Done | ${transformedTasks.filter((t) => t.status === "done").length} |`
        );
        lines.push("");

        // Task list
        lines.push(`## Tasks`);
        lines.push("");

        transformedTasks.forEach((task) => {
          const statusEmoji =
            task.status === "done"
              ? "✅"
              : task.status === "in-progress"
              ? "🔄"
              : "⏱️";
          const priorityBadge =
            task.priority === "high"
              ? "🔴"
              : task.priority === "medium"
              ? "🟡"
              : "🟢";

          lines.push(`### ${statusEmoji} ${task.title} ${priorityBadge}`);
          lines.push(`- **ID:** ${task.fullId}`);
          lines.push(`- **Status:** ${task.status}`);
          lines.push(`- **Priority:** ${task.priority}`);

          if (task.description) {
            lines.push("");
            lines.push(`**Description:** ${task.description}`);
          }

          if (task.details) {
            lines.push("");
            lines.push(`**Details:**`);
            lines.push("```");
            lines.push(task.details);
            lines.push("```");
          }

          if (task.subtasks.length > 0) {
            lines.push("");
            lines.push(
              `**Subtasks (${task.completedSubtasks}/${task.subtaskCount}):**`
            );
            task.subtasks.forEach((subtask) => {
              const subStatus =
                subtask.status === "done"
                  ? "✅"
                  : subtask.status === "in-progress"
                  ? "🔄"
                  : "⏱️";
              lines.push(`  - ${subStatus} \`${subtask.fullId}\` ${subtask.title}`);
            });
          }

          lines.push("");
        });

        return createContentResponse(lines.join("\n"));
      } catch (error) {
        log.error(`Error getting structured tasks: ${error.message}`);
        return createErrorResponse(
          `Error getting structured tasks: ${error.message}`
        );
      }
    },
  });
}
