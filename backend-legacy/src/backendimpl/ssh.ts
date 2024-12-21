import { NodeSSH } from "node-ssh";
import { Socket } from "node:net";

import type {
  BackendBaseClass,
  ForwardRule,
  ConnectedClient,
  ParameterReturnedValue,
} from "./base.js";

import {
  TcpConnectionDetails,
  AcceptConnection,
  ClientChannel,
  RejectConnection,
} from "ssh2";

type ForwardRuleExt = ForwardRule & {
  enabled: boolean;
};

// Fight me (for better naming)
type BackendParsedProviderString = {
  ip: string;
  port: number;

  username: string;
  privateKey: string;

  listenOnIPs: string[];
};

function parseBackendProviderString(data: string): BackendParsedProviderString {
  try {
    JSON.parse(data);
  } catch (e) {
    throw new Error("Payload body is not JSON");
  }

  const jsonData = JSON.parse(data);

  if (typeof jsonData.ip != "string") {
    throw new Error("IP field is not a string");
  }

  if (typeof jsonData.port != "number") {
    throw new Error("Port is not a number");
  }

  if (typeof jsonData.username != "string") {
    throw new Error("Username is not a string");
  }

  if (typeof jsonData.privateKey != "string") {
    throw new Error("Private key is not a string");
  }

  let listenOnIPs: string[] = [];

  if (!Array.isArray(jsonData.listenOnIPs)) {
    listenOnIPs.push("0.0.0.0");
  } else {
    listenOnIPs = jsonData.listenOnIPs;
  }

  return {
    ip: jsonData.ip,
    port: jsonData.port,

    username: jsonData.username,
    privateKey: jsonData.privateKey,

    listenOnIPs,
  };
}

export class SSHBackendProvider implements BackendBaseClass {
  state: "stopped" | "stopping" | "started" | "starting";

  clients: ConnectedClient[];
  proxies: ForwardRuleExt[];
  logs: string[];

  sshInstance: NodeSSH;
  options: BackendParsedProviderString;

  constructor(parameters: string) {
    this.logs = [];
    this.proxies = [];
    this.clients = [];

    this.options = parseBackendProviderString(parameters);

    this.state = "stopped";
  }

  async start(): Promise<boolean> {
    this.state = "starting";
    this.logs.push("Starting SSHBackendProvider...");

    if (this.sshInstance) {
      this.sshInstance.dispose();
    }

    this.sshInstance = new NodeSSH();

    try {
      await this.sshInstance.connect({
        host: this.options.ip,
        port: this.options.port,

        username: this.options.username,
        privateKey: this.options.privateKey,
      });
    } catch (e) {
      this.logs.push(`Failed to start SSHBackendProvider! Error: '${e}'`);
      this.state = "stopped";

      // @ts-expect-error: We know that stuff will be initialized in order, so this will be safe
      this.sshInstance = null;

      return false;
    }

    if (this.sshInstance.connection) {
      this.sshInstance.connection.on("end", async () => {
        if (this.state != "started") return;
        this.logs.push("We disconnected from the SSH server. Restarting...");

        // Create a new array from the existing list of proxies, so we have a backup of the proxy list before
        // we wipe the list of all proxies and clients (as we're disconnected anyways)
        const proxies = Array.from(this.proxies);

        this.proxies.splice(0, this.proxies.length);
        this.clients.splice(0, this.clients.length);

        await this.start();

        if (this.state != "started") return;

        for (const proxy of proxies) {
          if (!proxy.enabled) continue;

          this.addConnection(
            proxy.sourceIP,
            proxy.sourcePort,
            proxy.destPort,
            "tcp",
          );
        }
      });
    }

    this.state = "started";
    this.logs.push("Successfully started SSHBackendProvider.");

    return true;
  }

  async stop(): Promise<boolean> {
    this.state = "stopping";
    this.logs.push("Stopping SSHBackendProvider...");

    this.proxies.splice(0, this.proxies.length);

    this.sshInstance.dispose();

    // @ts-expect-error: We know that stuff will be initialized in order, so this will be safe
    this.sshInstance = null;

    this.logs.push("Successfully stopped SSHBackendProvider.");
    this.state = "stopped";

    return true;
  }

  addConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {
    const connectionCheck = SSHBackendProvider.checkParametersConnection(
      sourceIP,
      sourcePort,
      destPort,
      protocol,
    );

    if (!connectionCheck.success) throw new Error(connectionCheck.message);

    const foundProxyEntry = this.proxies.find(
      i =>
        i.sourceIP == sourceIP &&
        i.sourcePort == sourcePort &&
        i.destPort == destPort,
    );

    if (foundProxyEntry) return;

    const connCallback = (
      info: TcpConnectionDetails,
      accept: AcceptConnection<ClientChannel>,
      reject: RejectConnection,
    ) => {
      const foundProxyEntry = this.proxies.find(
        i =>
          i.sourceIP == sourceIP &&
          i.sourcePort == sourcePort &&
          i.destPort == destPort,
      );

      if (!foundProxyEntry || !foundProxyEntry.enabled) return reject();

      const client: ConnectedClient = {
        ip: info.srcIP,
        port: info.srcPort,

        connectionDetails: foundProxyEntry,
      };

      this.clients.push(client);

      const srcConn = new Socket();

      srcConn.connect({
        host: sourceIP,
        port: sourcePort,
      });

      // Why is this so confusing
      const destConn = accept();

      destConn.addListener("data", (chunk: Uint8Array) => {
        srcConn.write(chunk);
      });

      destConn.addListener("end", () => {
        this.clients.splice(this.clients.indexOf(client), 1);
        srcConn.end();
      });

      srcConn.on("data", data => {
        destConn.write(data);
      });

      srcConn.on("end", () => {
        this.clients.splice(this.clients.indexOf(client), 1);
        destConn.end();
      });
    };

    for (const ip of this.options.listenOnIPs) {
      this.sshInstance.forwardIn(ip, destPort, connCallback);
    }

    this.proxies.push({
      sourceIP,
      sourcePort,
      destPort,

      enabled: true,
    });
  }

  removeConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {
    const connectionCheck = SSHBackendProvider.checkParametersConnection(
      sourceIP,
      sourcePort,
      destPort,
      protocol,
    );

    if (!connectionCheck.success) throw new Error(connectionCheck.message);

    const foundProxyEntry = this.proxies.find(
      i =>
        i.sourceIP == sourceIP &&
        i.sourcePort == sourcePort &&
        i.destPort == destPort,
    );

    if (!foundProxyEntry) return;

    foundProxyEntry.enabled = false;
  }

  getAllConnections(): ConnectedClient[] {
    return this.clients;
  }

  static checkParametersConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): ParameterReturnedValue {
    if (protocol == "udp") {
      return {
        success: false,
        message:
          "SSH does not support UDP tunneling! Please use something like PortCopier instead (if it gets done)",
      };
    }

    return {
      success: true,
    };
  }

  static checkParametersBackendInstance(data: string): ParameterReturnedValue {
    try {
      parseBackendProviderString(data);
      // @ts-expect-error: We write the function, and we know we're returning an error
    } catch (e: Error) {
      return {
        success: false,
        message: e.toString(),
      };
    }

    return {
      success: true,
    };
  }
}
