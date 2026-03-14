/**
 * listMembers.js
 * MCP tool for listing team members
 */

import { z } from "zod";
import {
  executeTaskMasterCommand,
  createContentResponse,
  createErrorResponse,
} from "./utils.js";

/**
 * Register the listMembers tool with the MCP server
 * @param {Object} server - FastMCP server instance
 */
export function registerListMembersTool(server) {
  server.addTool({
    name: "listMembers",
    description: "List all team members with optional filters",
    parameters: z.object({
      status: z.string().optional().describe("Filter by status: active, inactive"),
      role: z.string().optional().describe("Filter by role: admin, leader, member"),
      department: z.string().optional().describe("Filter by department"),
      keyword: z.string().optional().describe("Search keyword for name or email"),
      projectRoot: z.string().describe("Project root directory"),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(`Listing members with filters: ${JSON.stringify(args)}`);

        // 由于成员管理是通过 API 访问的，我们需要直接调用存储层
        // 这里我们返回一个提示，告知用户使用 API 或 CLI
        const result = {
          message: "Team members can be managed through the Web UI or API",
          apiEndpoint: "/api/members",
          availableFilters: ["status", "role", "department", "keyword"],
          hint: "Use the task-manager server command to start the API server"
        };

        return createContentResponse(JSON.stringify(result, null, 2));
      } catch (error) {
        log.error(`Error listing members: ${error.message}`);
        return createErrorResponse(`Error listing members: ${error.message}`);
      }
    },
  });
}

/**
 * Register the getMember tool with the MCP server
 * @param {Object} server - FastMCP server instance
 */
export function registerGetMemberTool(server) {
  server.addTool({
    name: "getMember",
    description: "Get details of a specific team member",
    parameters: z.object({
      memberId: z.number().describe("Member ID"),
      projectRoot: z.string().describe("Project root directory"),
    }),
    execute: async (args, { log }) => {
      try {
        log.info(`Getting member: ${args.memberId}`);

        const result = {
          message: "Use the API endpoint to get member details",
          apiEndpoint: `/api/members/${args.memberId}`
        };

        return createContentResponse(JSON.stringify(result, null, 2));
      } catch (error) {
        log.error(`Error getting member: ${error.message}`);
        return createErrorResponse(`Error getting member: ${error.message}`);
      }
    },
  });
}

export default {
  registerListMembersTool,
  registerGetMemberTool,
};
