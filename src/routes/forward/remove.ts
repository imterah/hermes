import type { PrismaClient } from "@prisma/client";
import type { FastifyInstance } from "fastify";

import type { BackendInterface } from "../../backendimpl/base.js";
import { ServerOptions, SessionToken } from "../../libs/types.js";
import { hasPermissionByToken } from "../../libs/permissions.js";

export function route(fastify: FastifyInstance, prisma: PrismaClient, tokens: Record<number, SessionToken[]>, _options: ServerOptions, backends: Record<number, BackendInterface>) {
  function hasPermission(token: string, permissionList: string[]): Promise<boolean> {
    return hasPermissionByToken(permissionList, token, tokens, prisma);
  };

  /**
   * Creates a new route to use
   */
  fastify.post("/api/v1/forward/remove", {
    schema: {
      body: {
        type: "object",
        required: ["token", "id"],
  
        properties: {
          token: { type: "string" },
          id: { type: "number" }
        }
      }
    }
  }, async(req, res) => {
    // @ts-ignore
    const body: {
      token: string,
      id: number
    } = req.body;

    if (!await hasPermission(body.token, [
      "routes.remove"
    ])) {
      return res.status(403).send({
        error: "Unauthorized"
      });
    };
    
    await prisma.forwardRule.delete({
      where: {
        id: body.id
      }
    });

    return {
      success: true
    }
  });
}