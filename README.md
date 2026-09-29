# Twilio DR Agent Console

## Architecture

Current architecture:

```
Browser
  ↓
React Agent Console (Frontend)
  ↓
Go DR API (Backend)
  ↓
Firestore
```

Service layer architecture:

```
Go DR API
  ↓
Service layer
  ↓
CallService
  ↓
TelephonyProvider interface
  ↓
MockTelephonyProvider
  ↓
Mock Conference
```

Note: The `MockTelephonyProvider` will later be replaced with `TwilioTelephonyProvider` without changing the service layer. The current architecture keeps this intentionally simple.

## Environment variables

To run the application, you need to configure the environment variables. Copy `.env.example` to `.env` if you are using Docker, or export them in your shell:

- `BACKEND_PORT`: The port the Go backend listens on (default: `8080`).
- `FRONTEND_PORT`: The port the React frontend is exposed on in Docker (default: `3000`).
- `VITE_API_BASE_URL`: The URL the React frontend uses to reach the Go backend. **IMPORTANT:** This must be an address accessible from the browser (e.g., `http://localhost:8080`), not the internal Docker container name.

## Firestore credentials

The backend uses Google Application Default Credentials to connect to Firestore.

**For Local Development (without Docker):**
Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable pointing to your Firebase service account JSON file.

**For Docker Compose:**
Place your service account JSON file at `./secrets/firebase-service-account.json`. The docker-compose configuration will automatically mount this file into the backend container and set the appropriate environment variable.
*Note: Never commit your service account JSON file to version control. The `./secrets/` folder is included in `.gitignore`.*

## Running without Docker

1. **Start the Go Backend:**
   ```bash
   # Make sure GOOGLE_APPLICATION_CREDENTIALS is set
   cd twilio-go
   go run cmd/server/main.go
   ```

2. **Start the React Frontend:**
   ```bash
   cd twilio-go/frontend
   npm install
   npm run dev
   ```

## Running with Docker

You can run both the frontend and backend using Docker Compose.

```bash
docker compose up --build
```

The frontend will be available at `http://localhost:3000` (or whatever `FRONTEND_PORT` is set to).
The backend will be available at `http://localhost:8080` (or whatever `BACKEND_PORT` is set to).
You can verify the backend health at `http://localhost:8080/health`.

## Demo

The complete agent/call flow:
1. **Create agent**: Create an agent using the console.
2. **Enqueue call**: Use an API request (or UI if available) to enqueue a call.
3. **Queue appears**: The call appears in the agent's queue.
4. **Accept**: The agent accepts the call.
5. **Mock conference created**: A mock telephony conference is created.
6. **End call**: The call ends.
7. **Agent becomes available**: The agent state changes back to available.
# twilio-dr
