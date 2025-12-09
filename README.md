# 🐙 Webhook Gateway Project

This project provides a secure, decoupled, and highly reliable gateway for forwarding webhooks from a SaaS source (like GitHub Enterprise) through Azure to an on-premise tool.

## 🏗 Architecture

The system is decoupled into two primary microservices that communicate asynchronously via an Azure Service Bus (ASB) queue.

| Component | Deployment Location | Responsibility |
| :--- | :--- | :--- |
| **`receiver`** | AKS | Receives webhooks, validates the signature, responds immediately with `HTTP 200 OK`, and publishes the raw payload to ASB. |
| **`forwarder`** | AKS | Consumes messages from the ASB queue, extracts the payload, and forwards it via HTTP/S to the final on-premise tool endpoint. |

## 🚀 Getting Started

### Prerequisites

1.  Go (1.21+)
2.  Docker
3.  Kubernetes cluster (AKS)
4.  Azure Service Bus Queue
5.  GitHub Webhook Secret

### 1. Configuration (Environment Variables)

| Variable | Component | Description | Mandatory |
| :--- | :--- | :--- | :--- |
| `AZURE_SERVICE_BUS_CONN_STRING` | Both | Connection string for the Azure Service Bus namespace. | Yes |
| `AZURE_SERVICE_BUS_QUEUE_NAME` | Both | Name of the queue used for message transit. | Yes |
| `GITHUB_WEBHOOK_SECRET` | `receiver` | The shared secret used to validate the webhook signature. | Yes |
| `TARGET_TOOL_URL` | `forwarder` | Full HTTP/S URL of the final on-premise endpoint. | Yes |
| `TARGET_TOOL_AUTH_TOKEN` | `forwarder` | Authorization token for the on-premise tool. | No |

### 2. Build and Deploy

Use the `Dockerfile` in each `cmd/` subdirectory to build the respective services and deploy using the `k8s/` manifests for the `receiver`.
