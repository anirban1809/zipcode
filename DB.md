# DynamoDB Tables

This project interacts with several DynamoDB tables through the Lambda API. Each table is wired into the function via environment variables defined in the CDK stack (`lib/rest-api-stack.ts`).

| Table | Env Var | Description | Primary Keys | Notable Attributes |
| --- | --- | --- | --- | --- |
| `users_staging` | `USERS_TABLE_NAME` | Stores user profiles and workspace memberships. | PK = `USER`, email | `first_name`, `last_name`, `role`, `active`, `onboarding_required`, `onboarding_status`, `verified`, `workspaces` (map of workspaceId → role) |
| `workspaces_staging` | `WORKSPACES_TABLE_NAME` | Stores workspace metadata and membership map. | PK = `WORKSPACE`, SK = workspaceId | `workspace_name`, `owner_email`, `users` (map of email → role) |
| `invitations_staging` | `INVITATIONS_TABLE_NAME` | Tracks pending workspace invitations. | PK = `INVITATION`, SK = invitationId | `from`, `to`, `role`, `workspaceId`, `workspaceName`, `expiry`, `status` |
| `temp_staging` | `TEMP_TABLE_NAME` | Holds ephemeral key/value data for temporary/auxiliary workflows. | PK = provided `type`, SK = generated UUID | `data` (map with arbitrary payload) |
| `calendars_staging` | `CALENDARS_TABLE_NAME` | Persists calendar connections for workspace users. | PK = `WORKSPACE#${workspaceId}`, SK = `CALENDAR#${email}#${calendarId}` | `provider`, `calendar_id`, `user_email`, `owner_email`, `calendar_name`, `connected_at`, `updated_at`, `status` |
| `google_accounts_staging` | `GOOGLE_ACCOUNTS_TABLE` | Maps Google refresh tokens to workspace users. | PK = `WORKSPACE#${workspaceId}`, SK = `USER#${email}` | `owner`, `refresh_token`, `connected_at`, `updated_at`, `status` |

The CDK stack hard-codes the ARN references to these tables for IAM permissions so the Lambda can perform reads/writes. Most calls use the AWS SDK v3 DynamoDB client to `GetItem`, `PutItem`, `UpdateItem`, `BatchGetItem`, `Query`, and `DeleteItem` on the tables listed above. Parameter structures mirror the key schema in each helper function in `src/db/db.service.ts`.
