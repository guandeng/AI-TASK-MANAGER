/**
 * tools/getSubtask.js
 * Tool to get specific subtasks from another project for reference or reuse
 */

import { z } from "zod";
import {
  executeTaskMasterCommand,
  createContentResponse,
  createErrorResponse,
} from "./utils.js";
import fs from "fs";
import path from "path";

/**
 * Get task/subtask status information
 * @param {string} taskId - Task ID
 * @param {string} projectRoot - Project root directory
 * @returns {Object|null} - Status information or null
 */
function getTaskStatusInfo(taskId, projectRoot) {
  try {
    const tasksPath = path.join(projectRoot, "tasks", "tasks.json");
    if (!fs.existsSync(tasksPath)) {
      return null;
    }

    const tasksData = JSON.parse(fs.readFileSync(tasksPath, "utf8"));
    const [mainId, subId] = taskId.toString().split(".");

    const task = tasksData.tasks.find((t) => t.id == mainId);
    if (!task) return null;

    if (subId) {
      const subtask = task.subtasks?.find((st) => st.id == subId);
      if (!subtask) return null;
      return {
        status: subtask.status,
        title: subtask.title,
        type: "subtask",
        parentTitle: task.title,
        parentStatus: task.status,
      };
    }

    return {
      status: task.status,
      title: task.title,
      type: "task",
      subtaskCount: task.subtasks?.length || 0,
      completedSubtasks:
        task.subtasks?.filter((st) => st.status === "done").length || 0,
    };
  } catch (error) {
    return null;
  }
}

/**
 * Create status awareness message
 * @param {Object} statusInfo - Status information
 * @param {string} projectPath - Project path for context
 * @returns {string} - Status message
 */
function createStatusAwarenessMessage(statusInfo, projectPath) {
  const statusEmoji = {
    pending: "⏱️",
    "in-progress": "🔄",
    done: "✅",
    deferred: "📌",
    blocked: "🚫",
  };

  const emoji = statusEmoji[statusInfo.status] || "❓";
  const projectName = path.basename(projectPath);

  let message = `\n${emoji} **状态提示** (来源: ${projectName})\n`;

  if (statusInfo.type === "subtask") {
    message += `子任务 "${statusInfo.title}" (父任务: ${statusInfo.parentTitle}) 当前状态: **${statusInfo.status}**\n`;

    if (statusInfo.status !== "in-progress" && statusInfo.status !== "done") {
      message += `\n⚠️ **注意**: 此子任务尚未完成\n`;
      message += `- 父任务状态: ${statusInfo.parentStatus}\n`;
      message += `\n💡 **建议操作**:\n`;
      message += `1. 查看实现细节作为参考\n`;
      message += `2. 了解该子任务的测试策略\n`;
      message += `3. 在当前项目中创建类似任务\n`;
    } else if (statusInfo.status === "done") {
      message += `\n✅ **已完成**: 可以安全参考此实现\n`;
    }
  } else {
    message += `任务 "${statusInfo.title}" 当前状态: **${statusInfo.status}**\n`;

    if (statusInfo.subtaskCount > 0) {
      message += `进度: ${statusInfo.completedSubtasks}/${statusInfo.subtaskCount} 子任务已完成\n`;
    }

    if (statusInfo.status !== "in-progress" && statusInfo.status !== "done") {
      message += `\n⚠️ **注意**: 此任务尚未进行中\n`;
      message += `\n💡 **建议操作**:\n`;
      message += `1. 查看任务的子任务分解方式\n`;
      message += `2. 参考任务的实现策略\n`;
      message += `3. 了解任务的测试方法\n`;
    } else if (statusInfo.status === "done") {
      message += `\n✅ **已完成**: 所有子任务都已实现,可作为完整参考\n`;
    } else {
      message += `\n🔄 **进行中**: 部分功能已实现,可参考已完成部分\n`;
    }
  }

  message += `\n---\n`;

  return message;
}

/**
 * Register the getSubtask tool with the MCP server
 * @param {Object} server - FastMCP server instance
 */
export function registerGetSubtaskTool(server) {
  server.addTool({
    name: "getSubtask",
    description:
      "Get a specific subtask or all subtasks of a task from another project. Useful for reference, reuse, or understanding implementation details across projects. Includes status awareness to inform user about task progress.",
    parameters: z.object({
      sourceProject: z
        .string()
        .describe("Root directory of the source project to read subtask from"),
      taskId: z
        .string()
        .describe(
          "Task ID to get subtasks from (e.g., '1' for all subtasks of task 1, '1.2' for specific subtask)"
        ),
      includeDetails: z
        .boolean()
        .optional()
        .describe("Include full details (description, details, testStrategy)"),
      includeImplementation: z
        .boolean()
        .optional()
        .describe("Include implementation notes and code references"),
      format: z
        .enum(["summary", "detailed", "json"])
        .optional()
        .describe("Output format: summary (default), detailed, or json"),
      checkStatus: z
        .boolean()
        .optional()
        .describe(
          "Whether to check and inform about task status (default: true)"
        ),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(
          `Getting subtask ${args.taskId} from project ${args.sourceProject}`
        );

        // Check task status first (unless explicitly disabled)
        const shouldCheckStatus = args.checkStatus !== false;
        let statusMessage = "";

        if (shouldCheckStatus) {
          const statusInfo = getTaskStatusInfo(args.taskId, args.sourceProject);
          if (statusInfo) {
            statusMessage = createStatusAwarenessMessage(
              statusInfo,
              args.sourceProject
            );
          }
        }

        const result = executeTaskMasterCommand(
          "show",
          log,
          [args.taskId],
          args.sourceProject
        );

        if (!result.success) {
          throw new Error(result.error);
        }

        let output = result.stdout;

        // Format based on requested format
        if (args.format === "json") {
          try {
            // Try to parse and return as JSON if possible
            const tasksPath = path.join(args.sourceProject, "tasks", "tasks.json");
            const tasksData = JSON.parse(fs.readFileSync(tasksPath, "utf8"));

            // Find the specific task/subtask
            const [mainId, subId] = args.taskId.toString().split(".");

            const task = tasksData.tasks.find((t) => t.id == mainId);
            if (!task) {
              throw new Error(`Task ${mainId} not found`);
            }

            if (subId) {
              const subtask = task.subtasks?.find((st) => st.id == subId);
              if (!subtask) {
                throw new Error(
                  `Subtask ${args.taskId} not found in task ${mainId}`
                );
              }
              output = JSON.stringify(subtask, null, 2);
            } else {
              output = JSON.stringify(
                task.subtasks || { message: "No subtasks found" },
                null,
                2
              );
            }
          } catch (parseError) {
            log.warn(`Could not parse as JSON: ${parseError.message}`);
            // Fall back to text output
          }
        }

        // Add metadata header and status message
        const header = `📋 子任务来源: ${args.sourceProject}\n`;
        const separator = "─".repeat(60) + "\n";

        return createContentResponse(
          statusMessage + header + separator + output
        );
      } catch (error) {
        log.error(`Error getting subtask: ${error.message}`);
        return createErrorResponse(`Error getting subtask: ${error.message}`);
      }
    },
  });
}
