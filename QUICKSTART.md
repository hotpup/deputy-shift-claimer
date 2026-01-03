# Quick Start Guide

## Setup Steps

### 1. Install Python Dependencies
```bash
pip install -r requirements.txt
```

### 2. Configure Gmail API

1. Visit [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project
3. Enable Gmail API
4. Create OAuth 2.0 Desktop credentials
5. Download credentials as `credentials.json` in this directory

### 3. Configure Your Preferences

Copy the example config:
```bash
cp config.example.json config.json
```

Edit `config.json` to set your preferences:
- `target_shift_duration_hours`: Minimum hours for a shift to match
- `target_shift_roles`: List of role names to look for
- `gmail_label`: Gmail label for Deputy emails
- `notification_method`: "console" (more options coming soon)

### 4. Label Your Gmail Emails

In Gmail:
1. Create a label named "Deputy"
2. Apply it to Deputy shift notification emails
3. Optional: Set up a filter to auto-label incoming Deputy emails

### 5. Run the Script

```bash
python deputy_shift_claimer.py
```

On first run, you'll be prompted to authenticate with Google.

## Troubleshooting

**No messages found**: Make sure your Deputy emails have the correct label in Gmail.

**Authentication error**: Delete `token.json` and re-authenticate.

**Missing credentials**: Ensure `credentials.json` is in the same directory as the script.

## Customizing Shift Patterns

The script looks for common patterns in Deputy emails:
- "Shift: [Role Name]"
- "Position: [Role Name]"
- Time ranges like "9:00 AM - 5:00 PM"
- Explicit durations like "8 hours" or "8h"

If your Deputy emails have different formats, you may need to adjust the regex patterns in `deputy_shift_claimer.py`.
