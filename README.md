# Deputy Shift Claimer

A Go (Golang) application that parses Gmail emails with the "Deputy" label and checks for shift length and shift roles. It notifies you when shifts meet your target duration or role criteria.

## Features

- 🔍 **Email Parsing**: Automatically fetches and parses emails labeled "Deputy" from your Gmail
- ⏱️ **Shift Duration Detection**: Extracts shift duration from email content
- 👔 **Role Matching**: Identifies shift roles and matches against your target roles
- 🔔 **Smart Notifications**: Alerts you when shifts meet your criteria
- ⚙️ **Configurable**: Easy configuration via JSON file
- 🚀 **Fast & Efficient**: Built with Go for performance and reliability

## Prerequisites

- Go 1.21 or higher
- A Google Cloud Project with Gmail API enabled
- Gmail account with Deputy shift notification emails

## Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Set Up Gmail API Credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Gmail API for your project:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Gmail API"
   - Click "Enable"
4. Create OAuth 2.0 credentials:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth client ID"
   - Select "Desktop app" as the application type
   - Download the credentials and save as `credentials.json` in the project directory

### 3. Configure Your Preferences

Edit `config.json` to set your target shift criteria:

```json
{
  "target_shift_duration_hours": 8,
  "target_shift_roles": [
    "Bartender",
    "Server",
    "Manager"
  ],
  "gmail_label": "Deputy",
  "notification_method": "console"
}
```

**Configuration Options:**

- `target_shift_duration_hours`: Minimum shift duration in hours to match (default: 8)
- `target_shift_roles`: List of role names to match (case-insensitive partial matching)
- `gmail_label`: Gmail label to filter Deputy emails (default: "Deputy")
- `notification_method`: How to receive notifications (currently supports: "console")

### 4. Label Your Deputy Emails

In Gmail, create a label called "Deputy" (or use a custom name in config) and apply it to your Deputy shift notification emails. You can set up filters to automatically label incoming Deputy emails.

## Usage

### Build the Application

```bash
go build -o deputy-shift-claimer
```

### Run the Application

```bash
./deputy-shift-claimer
```

Or run directly without building:

```bash
go run main.go
```

On first run, the application will:
1. Open a browser window (or provide a URL) for Gmail authentication
2. Ask you to authorize the application
3. Save credentials to `token.json` for future use

The application will then:
1. Fetch all emails with the "Deputy" label
2. Parse each email for shift information
3. Check if shifts meet your target criteria
4. Notify you of any matches

## Example Output

```
Deputy Shift Claimer
============================================================
Target shift duration: >=8 hours
Target roles: Bartender, Server, Manager
Gmail label: Deputy
============================================================

Authenticating with Gmail...
✓ Successfully authenticated

Fetching emails with label 'Deputy'...
✓ Found 5 email(s)

Processing emails...

[1/5] Deputy Shift Available - Server Position...

============================================================
🎯 SHIFT MATCH FOUND!
============================================================
Email: Deputy Shift Available - Server Position
Date: Thu, 2 Jan 2026 10:30:00 -0800
Shift Role: Server
Duration: 8.0 hours
Start Time: 9:00 AM
End Time: 5:00 PM
Match Reason: Duration: 8.0h (target: >=8h); Role: Server (matches: Server)
============================================================

[2/5] Deputy Schedule Update...
...

============================================================
Processing complete!
Total emails processed: 5
Matching shifts found: 2
============================================================
```

## Running Tests

Run the test suite:

```bash
go test -v
```

Run tests with coverage:

```bash
go test -cover
```

## How It Works

The application uses several pattern matching techniques to extract shift information:

1. **Role Detection**: Looks for keywords like "Shift:", "Position:", "Role:" followed by the role name
2. **Time Range Parsing**: Detects time patterns like "9:00 AM - 5:00 PM" or "09:00-17:00"
3. **Duration Calculation**: Either calculates from time ranges or extracts explicit durations like "8 hours"

## Project Structure

```
.
├── main.go              # Main application code
├── main_test.go         # Unit tests
├── go.mod               # Go module definition
├── config.json          # Configuration file
├── config.example.json  # Example configuration
├── credentials.json     # OAuth credentials (not tracked)
├── token.json          # OAuth token (not tracked)
└── README.md           # This file
```

## Troubleshooting

### "credentials.json not found"
- Make sure you've downloaded OAuth credentials from Google Cloud Console
- Place the file in the same directory as the application

### "Label 'Deputy' not found in Gmail"
- Check that you've created the label in Gmail
- Verify the label name matches what's in `config.json`
- Label names are case-sensitive

### "No messages found"
- Make sure emails are labeled correctly in Gmail
- Check that you have Deputy emails in your inbox
- Try running with a different label name

### Authentication Issues
- Delete `token.json` and re-authenticate
- Check that Gmail API is enabled in Google Cloud Console
- Verify your OAuth consent screen is configured

### Build Issues
- Ensure Go 1.21 or higher is installed: `go version`
- Run `go mod tidy` to clean up dependencies
- Check for any missing dependencies: `go mod download`

## Python Version

A Python implementation is also available in this repository. See the Python files for that version.

## Future Enhancements

- Email notifications for matches
- Desktop notifications (Windows/Mac/Linux)
- Web dashboard for viewing matches
- Automatic shift claiming
- Support for multiple email providers
- Calendar integration
- Slack/Discord notifications

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.