/**
 * tools/taskContextTools.js
 * Tools for getting complete task context and updating code hints
 * These help AI understand the full picture for code generation
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
 * Get requirement data from storage
 * @param {string} projectRoot - Project root directory
 * @param {number} requirementId - Requirement ID
 * @returns {Object|null} - Requirement data or null
 */
function getRequirementData(projectRoot, requirementId) {
  try {
    const requirementsPath = path.join(projectRoot, "data", "requirements.json");
    if (!fs.existsSync(requirementsPath)) {
      return null;
    }

    const data = JSON.parse(fs.readFileSync(requirementsPath, "utf8"));
    return data.requirements?.find(r => r.id === requirementId) || null;
  } catch {
    return null;
  }
}

/**
 * Save task data to storage
 * @param {string} projectRoot - Project root directory
 * @param {Object} data - Task data to save
 */
function saveTaskData(projectRoot, data) {
  const tasksPath = path.join(projectRoot, "tasks", "tasks.json");
  if (fs.existsSync(tasksPath)) {
    fs.writeFileSync(tasksPath, JSON.stringify(data, null, 2), "utf8");
    return true;
  }

  const altTasksPath = path.join(projectRoot, "task.json");
  if (fs.existsSync(altTasksPath)) {
    fs.writeFileSync(altTasksPath, JSON.stringify(data, null, 2), "utf8");
    return true;
  }

  return false;
}

/**
 * Build complete task context for AI code generation
 * @param {Object} task - Task object
 * @param {Object} requirement - Parent requirement (if any)
 * @param {Array} allTasks - All tasks for dependency resolution
 * @returns {Object} - Complete context
 */
function buildTaskContext(task, requirement, allTasks) {
  const context = {
    // 当前任务
    currentTask: {
      id: task.id,
      title: task.title || task.titleTrans,
      description: task.description || task.descriptionTrans,
      details: task.details || task.detailsTrans,
      testStrategy: task.testStrategy || task.testStrategyTrans,
      status: task.status,
      priority: task.priority,
      techStack: task.techStack || [],
      relatedFiles: task.relatedFiles || [],
      codeHints: task.codeHints || "",
      executionOrder: task.executionOrder,
    },

    // 父需求（如果有）
    parentRequirement: null,

    // 依赖任务
    dependencies: [],

    // 被依赖的任务
    dependents: [],

    // 子任务结构
    subtasks: [],

    // AI 代码生成建议
    codeGenerationHints: {
      suggestedFiles: [],
      suggestedInterfaces: [],
      implementationOrder: [],
    },
  };

  // 添加父需求信息
  if (requirement) {
    context.parentRequirement = {
      id: requirement.id,
      title: requirement.title,
      status: requirement.status,
      priority: requirement.priority,
      techStack: requirement.techStack || [],
    };
  }

  // 解析依赖任务
  if (task.dependencies && task.dependencies.length > 0) {
    context.dependencies = task.dependencies.map(depId => {
      const depTask = allTasks.find(t => t.id === depId);
      if (depTask) {
        return {
          id: depTask.id,
          title: depTask.title || depTask.titleTrans,
          status: depTask.status,
          relatedFiles: depTask.relatedFiles || [],
        };
      }
      return { id: depId, title: "Unknown", status: "unknown" };
    });
  }

  // 查找被依赖的任务
  context.dependents = allTasks
    .filter(t => t.dependencies && t.dependencies.includes(task.id))
    .map(t => ({
      id: t.id,
      title: t.title || t.titleTrans,
      status: t.status,
    }));

  // 处理子任务
  if (task.subtasks && task.subtasks.length > 0) {
    context.subtasks = task.subtasks.map((subtask, index) => ({
      id: `${task.id}.${subtask.id || index + 1}`,
      title: subtask.title || subtask.titleTrans,
      description: subtask.description || subtask.descriptionTrans,
      details: subtask.details || subtask.detailsTrans,
      status: subtask.status,
      relatedFiles: subtask.relatedFiles || [],
      codeInterface: subtask.codeInterface || null,
      acceptanceCriteria: subtask.acceptanceCriteria || [],
      codeHints: subtask.codeHints || "",
      executionOrder: subtask.executionOrder,
      dependencies: subtask.dependencies || [],
    }));

    // 生成实现顺序建议
    context.codeGenerationHints.implementationOrder = context.subtasks
      .filter(s => s.status !== "done")
      .sort((a, b) => (a.executionOrder || 999) - (b.executionOrder || 999))
      .map(s => ({
        id: s.id,
        title: s.title,
        files: s.relatedFiles,
      }));
  }

  // 汇总建议的文件
  const allFiles = new Set();
  if (task.relatedFiles) {
    task.relatedFiles.forEach(f => allFiles.add(f));
  }
  if (task.subtasks) {
    task.subtasks.forEach(s => {
      if (s.relatedFiles) {
        s.relatedFiles.forEach(f => allFiles.add(f));
      }
    });
  }
  context.codeGenerationHints.suggestedFiles = Array.from(allFiles);

  return context;
}

/**
 * Generate AI prompt from task context
 * @param {Object} context - Task context
 * @returns {string} - AI-optimized prompt
 */
function generateContextPrompt(context) {
  const lines = [];

  lines.push(`# Task Context for AI Code Generation`);
  lines.push(`Generated: ${new Date().toISOString()}`);
  lines.push("");

  // 当前任务
  lines.push(`## Current Task: ${context.currentTask.title}`);
  lines.push(`- **ID:** ${context.currentTask.id}`);
  lines.push(`- **Status:** ${context.currentTask.status}`);
  lines.push(`- **Priority:** ${context.currentTask.priority}`);

  if (context.currentTask.techStack?.length > 0) {
    lines.push(`- **Tech Stack:** ${context.currentTask.techStack.join(", ")}`);
  }

  if (context.currentTask.description) {
    lines.push(`\n### Description\n${context.currentTask.description}`);
  }

  if (context.currentTask.details) {
    lines.push(`\n### Implementation Details\n${context.currentTask.details}`);
  }

  if (context.currentTask.testStrategy) {
    lines.push(`\n### Test Strategy\n${context.currentTask.testStrategy}`);
  }

  if (context.currentTask.codeHints) {
    lines.push(`\n### Existing Code Hints\n\`\`\`\n${context.currentTask.codeHints}\n\`\`\``);
  }

  if (context.currentTask.relatedFiles?.length > 0) {
    lines.push(`\n### Related Files`);
    context.currentTask.relatedFiles.forEach(f => {
      lines.push(`- \`${f}\``);
    });
  }

  // 父需求
  if (context.parentRequirement) {
    lines.push(`\n## Parent Requirement: ${context.parentRequirement.title}`);
    lines.push(`- **ID:** ${context.parentRequirement.id}`);
    lines.push(`- **Status:** ${context.parentRequirement.status}`);
    if (context.parentRequirement.techStack?.length > 0) {
      lines.push(`- **Tech Stack:** ${context.parentRequirement.techStack.join(", ")}`);
    }
  }

  // 依赖任务
  if (context.dependencies.length > 0) {
    lines.push(`\n## Dependencies (Must Complete First)`);
    context.dependencies.forEach(dep => {
      const statusIcon = dep.status === "done" ? "✅" : "⏱️";
      lines.push(`- ${statusIcon} **${dep.id}**: ${dep.title}`);
      if (dep.relatedFiles?.length > 0) {
        lines.push(`  - Files: ${dep.relatedFiles.map(f => `\`${f}\``).join(", ")}`);
      }
    });
  }

  // 被依赖的任务
  if (context.dependents.length > 0) {
    lines.push(`\n## Dependents (Will Use This Task)`);
    context.dependents.forEach(dep => {
      lines.push(`- **${dep.id}**: ${dep.title} (${dep.status})`);
    });
  }

  // 子任务结构
  if (context.subtasks.length > 0) {
    const completedCount = context.subtasks.filter(s => s.status === "done").length;
    lines.push(`\n## Subtasks (${completedCount}/${context.subtasks.length} completed)`);

    context.subtasks.forEach(subtask => {
      const statusIcon = subtask.status === "done" ? "✅" : subtask.status === "in-progress" ? "🔄" : "⏱️";
      lines.push(`\n### ${statusIcon} ${subtask.id}: ${subtask.title}`);
      lines.push(`- **Status:** ${subtask.status}`);

      if (subtask.description) {
        lines.push(`- **Description:** ${subtask.description}`);
      }

      if (subtask.relatedFiles?.length > 0) {
        lines.push(`- **Files:** ${subtask.relatedFiles.map(f => `\`${f}\``).join(", ")}`);
      }

      if (subtask.codeInterface) {
        lines.push(`- **Interface:**`);
        lines.push(`  - Name: ${subtask.codeInterface.name}`);
        if (subtask.codeInterface.inputs) {
          lines.push(`  - Inputs: ${subtask.codeInterface.inputs}`);
        }
        if (subtask.codeInterface.outputs) {
          lines.push(`  - Outputs: ${subtask.codeInterface.outputs}`);
        }
        if (subtask.codeInterface.example) {
          lines.push(`  - Example: \`${subtask.codeInterface.example}\``);
        }
      }

      if (subtask.acceptanceCriteria?.length > 0) {
        lines.push(`- **Acceptance Criteria:**`);
        subtask.acceptanceCriteria.forEach((criteria, idx) => {
          const check = criteria.completed ? "x" : " ";
          lines.push(`  - [${check}] ${criteria.description}`);
        });
      }

      if (subtask.codeHints) {
        lines.push(`- **Code Hints:**`);
        lines.push(`\`\`\``);
        lines.push(subtask.codeHints);
        lines.push(`\`\`\``);
      }
    });
  }

  // 代码生成建议
  lines.push(`\n## Code Generation Hints`);

  if (context.codeGenerationHints.suggestedFiles.length > 0) {
    lines.push(`\n### Suggested Files to Create/Modify`);
    context.codeGenerationHints.suggestedFiles.forEach(f => {
      lines.push(`- \`${f}\``);
    });
  }

  if (context.codeGenerationHints.implementationOrder.length > 0) {
    lines.push(`\n### Recommended Implementation Order`);
    context.codeGenerationHints.implementationOrder.forEach((item, idx) => {
      lines.push(`${idx + 1}. **${item.id}**: ${item.title}`);
      if (item.files?.length > 0) {
        lines.push(`   - Files: ${item.files.map(f => `\`${f}\``).join(", ")}`);
      }
    });
  }

  // AI 指令
  lines.push(`\n---`);
  lines.push(`\n## Instructions for AI`);
  lines.push(`1. Review the task context and understand the requirements`);
  lines.push(`2. Check dependencies and ensure they are completed`);
  lines.push(`3. Follow the recommended implementation order for subtasks`);
  lines.push(`4. Use the code interfaces and acceptance criteria as guidance`);
  lines.push(`5. Create or modify files as suggested`);
  lines.push(`6. Ensure code quality and test coverage`);

  return lines.join("\n");
}

/**
 * Register the getTaskContext tool
 * @param {Object} server - FastMCP server instance
 */
export function registerGetTaskContextTool(server) {
  server.addTool({
    name: "getTaskContext",
    description:
      "Get complete task context for AI code generation. Includes task details, parent requirement, dependencies, subtasks structure, code interfaces, acceptance criteria, and suggested implementation order. Best for understanding the full picture before generating code.",
    parameters: z.object({
      projectRoot: z
        .string()
        .describe("Root directory of the AI-TASK-MANAGER project"),
      taskId: z
        .string()
        .describe("Task ID to get context for (e.g., '1')"),
      format: z
        .enum(["json", "prompt"])
        .optional()
        .default("json")
        .describe("Output format: json (machine-readable) or prompt (AI-optimized text)"),
      includeCodeHints: z
        .boolean()
        .optional()
        .default(true)
        .describe("Include existing code hints in the context"),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(`Getting task context for ID: ${args.taskId}`);

        const data = getTaskData(args.projectRoot);
        if (!data) {
          return createErrorResponse(
            `No task data found in ${args.projectRoot}`
          );
        }

        const tasks = data.tasks || [];
        const task = tasks.find(t => String(t.id) === args.taskId);

        if (!task) {
          return createErrorResponse(`Task not found: ${args.taskId}`);
        }

        // Get parent requirement if exists
        let requirement = null;
        if (task.requirementId) {
          requirement = getRequirementData(args.projectRoot, task.requirementId);
        }

        // Build complete context
        const context = buildTaskContext(task, requirement, tasks);

        // Remove code hints if not requested
        if (!args.includeCodeHints) {
          context.currentTask.codeHints = "";
          context.subtasks.forEach(s => s.codeHints = "");
        }

        if (args.format === "prompt") {
          return createContentResponse(generateContextPrompt(context));
        }

        return createContentResponse(JSON.stringify(context, null, 2));
      } catch (error) {
        log.error(`Error getting task context: ${error.message}`);
        return createErrorResponse(`Error getting task context: ${error.message}`);
      }
    },
  });
}

/**
 * Register the updateTaskCodeHints tool
 * @param {Object} server - FastMCP server instance
 */
export function registerUpdateTaskCodeHintsTool(server) {
  server.addTool({
    name: "updateTaskCodeHints",
    description:
      "Update code hints for a task or subtask. AI can use this to write implementation thoughts, code snippets, or technical decisions for later reference. This helps maintain context and guide future code generation.",
    parameters: z.object({
      projectRoot: z
        .string()
        .describe("Root directory of the AI-TASK-MANAGER project"),
      taskId: z
        .string()
        .describe("Task ID (e.g., '1') or subtask ID (e.g., '1.2')"),
      codeHints: z
        .string()
        .describe("Code hints content to save (markdown format supported)"),
      append: z
        .boolean()
        .optional()
        .default(false)
        .describe("Append to existing hints instead of replacing"),
      relatedFiles: z
        .array(z.string())
        .optional()
        .describe("List of related source files"),
      codeInterface: z
        .object({
          name: z.string(),
          inputs: z.string().optional(),
          outputs: z.string().optional(),
          example: z.string().optional(),
        })
        .optional()
        .describe("Code interface definition (for subtasks only)"),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(`Updating code hints for task: ${args.taskId}`);

        const data = getTaskData(args.projectRoot);
        if (!data) {
          return createErrorResponse(
            `No task data found in ${args.projectRoot}`
          );
        }

        const [mainId, subId] = args.taskId.toString().split(".");
        const taskIndex = data.tasks.findIndex(t => String(t.id) === mainId);

        if (taskIndex === -1) {
          return createErrorResponse(`Task not found: ${mainId}`);
        }

        const task = data.tasks[taskIndex];

        if (subId) {
          // Update subtask
          const subtaskIndex = task.subtasks?.findIndex(s => String(s.id) === subId);
          if (subtaskIndex === -1 || subtaskIndex === undefined) {
            return createErrorResponse(`Subtask not found: ${args.taskId}`);
          }

          if (!task.subtasks[subtaskIndex]) {
            task.subtasks[subtaskIndex] = {};
          }

          const subtask = task.subtasks[subtaskIndex];

          // Update code hints
          if (args.append && subtask.codeHints) {
            subtask.codeHints = subtask.codeHints + "\n\n" + args.codeHints;
          } else {
            subtask.codeHints = args.codeHints;
          }

          // Update related files if provided
          if (args.relatedFiles) {
            subtask.relatedFiles = args.relatedFiles;
          }

          // Update code interface if provided
          if (args.codeInterface) {
            subtask.codeInterface = args.codeInterface;
          }
        } else {
          // Update main task
          if (args.append && task.codeHints) {
            task.codeHints = task.codeHints + "\n\n" + args.codeHints;
          } else {
            task.codeHints = args.codeHints;
          }

          if (args.relatedFiles) {
            task.relatedFiles = args.relatedFiles;
          }
        }

        // Save updated data
        const saved = saveTaskData(args.projectRoot, data);
        if (!saved) {
          return createErrorResponse("Failed to save task data");
        }

        return createContentResponse(
          JSON.stringify({
            success: true,
            message: `Code hints updated for task ${args.taskId}`,
            taskId: args.taskId,
            timestamp: new Date().toISOString(),
          }, null, 2)
        );
      } catch (error) {
        log.error(`Error updating code hints: ${error.message}`);
        return createErrorResponse(`Error updating code hints: ${error.message}`);
      }
    },
  });
}

/**
 * Register the getNextTaskToImplement tool
 * @param {Object} server - FastMCP server instance
 */
export function registerGetNextTaskToImplementTool(server) {
  server.addTool({
    name: "getNextTaskToImplement",
    description:
      "Get the next recommended task/subtask to implement. Analyzes task dependencies, priorities, and current status to suggest the best next step for code generation.",
    parameters: z.object({
      projectRoot: z
        .string()
        .describe("Root directory of the AI-TASK-MANAGER project"),
      requirementId: z
        .string()
        .optional()
        .describe("Filter by requirement ID"),
      preferHighPriority: z
        .boolean()
        .optional()
        .default(true)
        .describe("Prefer high priority tasks"),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(`Getting next task to implement`);

        const data = getTaskData(args.projectRoot);
        if (!data) {
          return createErrorResponse(
            `No task data found in ${args.projectRoot}`
          );
        }

        let tasks = data.tasks || [];

        // Filter by requirement if specified
        if (args.requirementId) {
          tasks = tasks.filter(t => String(t.requirementId) === args.requirementId);
        }

        // Filter to only pending/in-progress tasks
        const activeTasks = tasks.filter(t => t.status === "pending" || t.status === "in-progress");

        if (activeTasks.length === 0) {
          return createContentResponse(
            JSON.stringify({
              message: "No pending tasks found. All tasks are completed or deferred.",
              recommendation: null,
            }, null, 2)
          );
        }

        // Find tasks with all dependencies completed
        const readyTasks = activeTasks.filter(task => {
          if (!task.dependencies || task.dependencies.length === 0) {
            return true;
          }
          return task.dependencies.every(depId => {
            const depTask = tasks.find(t => t.id === depId);
            return depTask && depTask.status === "done";
          });
        });

        if (readyTasks.length === 0) {
          // All tasks have unmet dependencies
          const blockedBy = activeTasks.map(task => ({
            id: task.id,
            title: task.title || task.titleTrans,
            blockedBy: (task.dependencies || [])
              .map(depId => {
                const dep = tasks.find(t => t.id === depId);
                return dep ? { id: depId, title: dep.title, status: dep.status } : null;
              })
              .filter(Boolean),
          }));

          return createContentResponse(
            JSON.stringify({
              message: "All active tasks have unmet dependencies",
              blockedTasks: blockedBy,
              recommendation: "Complete the blocking tasks first",
            }, null, 2)
          );
        }

        // Sort by priority and find the best task
        const priorityWeight = { high: 3, medium: 2, low: 1 };
        readyTasks.sort((a, b) => {
          // Prefer in-progress tasks first
          if (a.status === "in-progress" && b.status !== "in-progress") return -1;
          if (b.status === "in-progress" && a.status !== "in-progress") return 1;

          // Then by priority
          const aPriority = priorityWeight[a.priority] || 2;
          const bPriority = priorityWeight[b.priority] || 2;
          if (args.preferHighPriority) {
            return bPriority - aPriority;
          }
          return aPriority - bPriority;
        });

        const recommendedTask = readyTasks[0];

        // Find the next subtask to implement
        let nextSubtask = null;
        if (recommendedTask.subtasks && recommendedTask.subtasks.length > 0) {
          // Find pending subtasks with completed dependencies
          const readySubtasks = recommendedTask.subtasks.filter(st => {
            if (st.status === "done") return false;
            if (!st.dependencies || st.dependencies.length === 0) return true;
            return st.dependencies.every(depId => {
              const dep = recommendedTask.subtasks.find(s => s.id === depId);
              return dep && dep.status === "done";
            });
          });

          if (readySubtasks.length > 0) {
            // Sort by execution order or priority
            readySubtasks.sort((a, b) => (a.executionOrder || 999) - (b.executionOrder || 999));
            nextSubtask = readySubtasks[0];
          }
        }

        // Build response
        const response = {
          recommendation: {
            taskId: recommendedTask.id,
            title: recommendedTask.title || recommendedTask.titleTrans,
            status: recommendedTask.status,
            priority: recommendedTask.priority,
            subtaskCount: recommendedTask.subtasks?.length || 0,
            completedSubtasks: recommendedTask.subtasks?.filter(s => s.status === "done").length || 0,
            nextSubtask: nextSubtask ? {
              id: `${recommendedTask.id}.${nextSubtask.id}`,
              title: nextSubtask.title || nextSubtask.titleTrans,
              status: nextSubtask.status,
              relatedFiles: nextSubtask.relatedFiles || [],
              codeInterface: nextSubtask.codeInterface || null,
            } : null,
          },
          reasoning: recommendedTask.status === "in-progress"
            ? "This task is already in progress"
            : `This task has ${recommendedTask.priority} priority and all dependencies are met`,
          availableTasks: readyTasks.length,
          blockedTasks: activeTasks.length - readyTasks.length,
        };

        return createContentResponse(JSON.stringify(response, null, 2));
      } catch (error) {
        log.error(`Error getting next task: ${error.message}`);
        return createErrorResponse(`Error getting next task: ${error.message}`);
      }
    },
  });
}

export default {
  registerGetTaskContextTool,
  registerUpdateTaskCodeHintsTool,
  registerGetNextTaskToImplementTool,
};
