# deputy-shift-claimer

A Gmail parser for Deputy shift notifications written in Go. This tool monitors your Gmail inbox for Deputy shift notifications (in the "Deputy" label) and notifies you when shifts meet specific criteria.

## Features

- Fetches emails from Gmail "Deputy" label
- Parses Deputy shift notification emails
- Filters shifts based on:
  - **Role**: Notifies for specific roles (LG: ALL, LG: North, Deck Coordinator)
  - **Duration**: Notifies when shift duration meets or exceeds a threshold
- Configurable minimum duration threshold
- OAuth2 authentication with Gmail API

## Prerequisites

- Go 1.24 or later
- Gmail account
- Google Cloud Project with Gmail API enabled
- OAuth2 credentials

## Setup

### 1. Enable Gmail API

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Gmail API for your project
4. Go to "Credentials" and create OAuth2 credentials (Desktop application)
5. Download the credentials JSON file and save it as `credentials.json` in the project root

### 2. Install Dependencies

```bash
go mod download
```

### 3. Build the Application

```bash
go build -o deputy-shift-claimer
```

## Configuration

The application uses environment variables for configuration:

- `MIN_DURATION_HOURS`: Minimum shift duration in hours to trigger notification (default: 4.0)

Example:
```bash
export MIN_DURATION_HOURS=6.0
./deputy-shift-claimer
```

### Allowed Roles

By default, the application notifies for these roles:
- LG: ALL
- LG: North
- Deck Coordinator

To modify the allowed roles, edit the `main.go` file and update the `AllowedRoles` array in the `FilterConfig`.

## Usage

### First Run

On the first run, the application will:
1. Open a browser window for Google OAuth2 authentication
2. Ask you to authorize the application to read your Gmail
3. Save the authentication token to `token.json`

```bash
./deputy-shift-claimer
```

### Subsequent Runs

The application will use the saved token from `token.json`:

```bash
./deputy-shift-claimer
```

### Example Output

```
NOTIFICATION: Shift available!
  Role: LG: ALL
  Duration: 6.00 hours
  Date: 01/15/2026
  Time: 9:00 AM - 3:00 PM
  Message ID: abc123def456

NOTIFICATION: Shift available!
  Role: Deck Coordinator
  Duration: 4.50 hours
  Date: 01/16/2026
  Time: 2:00 PM - 6:30 PM
  Message ID: xyz789ghi012
```

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

## Project Structure

```
deputy-shift-claimer/
├── main.go                      # Application entry point
├── internal/
│   ├── gmail/
│   │   └── client.go           # Gmail API client
│   └── parser/
│       ├── parser.go           # Shift parser and filter logic
│       └── parser_test.go      # Unit tests
├── credentials.json            # OAuth2 credentials (not committed)
├── token.json                  # OAuth2 token (not committed)
├── go.mod                      # Go module definition
└── README.md                   # This file
```

## How It Works

1. **Gmail Client**: Connects to Gmail API using OAuth2 and fetches emails from the "Deputy" label
2. **Parser**: Extracts shift information from email content:
   - Role/Position
   - Date
   - Start and end times
   - Calculates duration in hours
3. **Filter**: Determines if notification should be sent based on:
   - Role matches one of the allowed roles OR
   - Duration meets or exceeds the minimum threshold

## Security

- `credentials.json` and `token.json` should never be committed to version control
- Add them to your `.gitignore` file
- The application uses OAuth2 with read-only Gmail scope (`gmail.readonly`)

## Troubleshooting

### "unable to read credentials file" error

Make sure `credentials.json` exists in the project root directory.

### Authentication issues

Delete `token.json` and run the application again to re-authenticate.

### No shifts found

- Verify that Deputy emails are labeled with "Deputy" in Gmail
- Check that the email format matches the parsing patterns

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
