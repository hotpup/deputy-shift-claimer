#!/usr/bin/env python3
"""
Deputy Shift Claimer
A script to parse Gmail emails with the "Deputy" label and check for shift
length and shift roles, notifying when target criteria are met.
"""

import os
import json
import re
from datetime import datetime
from typing import List, Dict, Optional, Tuple

from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError
import base64
from email.mime.text import MIMEText

# If modifying these scopes, delete the file token.json.
SCOPES = ['https://www.googleapis.com/auth/gmail.readonly']


def load_config(config_path: str = 'config.json') -> Dict:
    """Load configuration from JSON file."""
    with open(config_path, 'r') as f:
        return json.load(f)


def authenticate_gmail() -> Optional[object]:
    """Authenticate with Gmail API and return the service object."""
    creds = None
    
    # The file token.json stores the user's access and refresh tokens
    if os.path.exists('token.json'):
        creds = Credentials.from_authorized_user_file('token.json', SCOPES)
    
    # If there are no (valid) credentials available, let the user log in
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            if not os.path.exists('credentials.json'):
                print("Error: credentials.json not found!")
                print("Please download OAuth credentials from Google Cloud Console")
                return None
            
            flow = InstalledAppFlow.from_client_secrets_file(
                'credentials.json', SCOPES)
            creds = flow.run_local_server(port=0)
        
        # Save the credentials for the next run
        with open('token.json', 'w') as token:
            token.write(creds.to_json())
    
    try:
        service = build('gmail', 'v1', credentials=creds)
        return service
    except HttpError as error:
        print(f'An error occurred: {error}')
        return None


def get_deputy_emails(service: object, label: str = 'Deputy') -> List[Dict]:
    """Fetch emails with the specified label from Gmail."""
    try:
        # Get the label ID
        labels_result = service.users().labels().list(userId='me').execute()
        labels = labels_result.get('labels', [])
        
        label_id = None
        for lbl in labels:
            if lbl['name'] == label:
                label_id = lbl['id']
                break
        
        if not label_id:
            print(f"Label '{label}' not found in Gmail")
            return []
        
        # Fetch messages with the label
        results = service.users().messages().list(
            userId='me',
            labelIds=[label_id],
            maxResults=50  # Adjust as needed
        ).execute()
        
        messages = results.get('messages', [])
        
        if not messages:
            print(f"No messages found with label '{label}'")
            return []
        
        # Fetch full message details
        detailed_messages = []
        for message in messages:
            msg = service.users().messages().get(
                userId='me',
                id=message['id'],
                format='full'
            ).execute()
            detailed_messages.append(msg)
        
        return detailed_messages
    
    except HttpError as error:
        print(f'An error occurred: {error}')
        return []


def decode_email_body(payload: Dict) -> str:
    """Decode the email body from the payload."""
    body = ""
    
    if 'parts' in payload:
        # Multipart message
        for part in payload['parts']:
            if part['mimeType'] == 'text/plain':
                if 'data' in part['body']:
                    body = base64.urlsafe_b64decode(
                        part['body']['data']
                    ).decode('utf-8')
                    break
            elif part['mimeType'] == 'text/html' and not body:
                if 'data' in part['body']:
                    body = base64.urlsafe_b64decode(
                        part['body']['data']
                    ).decode('utf-8')
    else:
        # Simple message
        if 'data' in payload['body']:
            body = base64.urlsafe_b64decode(
                payload['body']['data']
            ).decode('utf-8')
    
    return body


def extract_shift_info(email_body: str, subject: str) -> Optional[Dict]:
    """
    Extract shift information from Deputy email.
    
    Returns a dictionary with:
    - role: The shift role/position
    - duration_hours: Duration in hours
    - start_time: Start time (if available)
    - end_time: End time (if available)
    """
    shift_info = {
        'role': None,
        'duration_hours': None,
        'start_time': None,
        'end_time': None
    }
    
    # Common Deputy email patterns
    # Pattern 1: "Shift: [Role] - [Date/Time]"
    role_pattern = r'(?:Shift|Position|Role):\s*([A-Za-z\s]+?)(?:\s*-|\s*\n|$)'
    role_match = re.search(role_pattern, email_body, re.IGNORECASE)
    if not role_match:
        role_match = re.search(role_pattern, subject, re.IGNORECASE)
    
    if role_match:
        shift_info['role'] = role_match.group(1).strip()
    
    # Pattern 2: Time ranges like "9:00 AM - 5:00 PM" or "09:00-17:00"
    time_pattern = r'(\d{1,2}:\d{2}\s*(?:AM|PM|am|pm)?)\s*[-–—to]+\s*(\d{1,2}:\d{2}\s*(?:AM|PM|am|pm)?)'
    time_match = re.search(time_pattern, email_body)
    
    if time_match:
        shift_info['start_time'] = time_match.group(1).strip()
        shift_info['end_time'] = time_match.group(2).strip()
        
        # Calculate duration
        try:
            duration = calculate_duration(
                shift_info['start_time'],
                shift_info['end_time']
            )
            shift_info['duration_hours'] = duration
        except Exception:
            pass
    
    # Pattern 3: Explicit duration like "8 hours" or "8h"
    duration_pattern = r'(\d+(?:\.\d+)?)\s*(?:hours?|hrs?|h)\b'
    duration_match = re.search(duration_pattern, email_body, re.IGNORECASE)
    
    if duration_match and not shift_info['duration_hours']:
        shift_info['duration_hours'] = float(duration_match.group(1))
    
    return shift_info if (shift_info['role'] or shift_info['duration_hours']) else None


def calculate_duration(start_time: str, end_time: str) -> float:
    """Calculate duration in hours between start and end times."""
    from dateutil import parser
    
    # Parse times
    start = parser.parse(start_time)
    end = parser.parse(end_time)
    
    # If end time is earlier than start time, assume it's next day
    if end < start:
        from datetime import timedelta
        end = end + timedelta(days=1)
    
    # Calculate duration
    duration = (end - start).total_seconds() / 3600
    return duration


def check_criteria(
    shift_info: Dict,
    target_duration: float,
    target_roles: List[str]
) -> Tuple[bool, str]:
    """
    Check if shift meets target criteria.
    
    Returns:
    - (True, reason) if criteria are met
    - (False, "") if criteria are not met
    """
    reasons = []
    
    # Check duration
    if shift_info.get('duration_hours'):
        if shift_info['duration_hours'] >= target_duration:
            reasons.append(
                f"Duration: {shift_info['duration_hours']}h "
                f"(target: >={target_duration}h)"
            )
    
    # Check role
    if shift_info.get('role'):
        for target_role in target_roles:
            if target_role.lower() in shift_info['role'].lower():
                reasons.append(f"Role: {shift_info['role']} (matches: {target_role})")
                break
    
    if reasons:
        return True, "; ".join(reasons)
    
    return False, ""


def notify(message: str, method: str = 'console'):
    """Send notification using the specified method."""
    if method == 'console':
        print("\n" + "="*60)
        print("🎯 SHIFT MATCH FOUND!")
        print("="*60)
        print(message)
        print("="*60 + "\n")
    else:
        # Future: Add email, desktop notification, etc.
        print(message)


def get_email_subject(message: Dict) -> str:
    """Extract subject from email message."""
    headers = message['payload']['headers']
    for header in headers:
        if header['name'].lower() == 'subject':
            return header['value']
    return ""


def get_email_date(message: Dict) -> str:
    """Extract date from email message."""
    headers = message['payload']['headers']
    for header in headers:
        if header['name'].lower() == 'date':
            return header['value']
    return ""


def main():
    """Main function to run the Deputy shift claimer."""
    print("Deputy Shift Claimer")
    print("=" * 60)
    
    # Load configuration
    try:
        config = load_config()
    except FileNotFoundError:
        print("Error: config.json not found!")
        return
    except json.JSONDecodeError:
        print("Error: config.json is not valid JSON!")
        return
    
    target_duration = config.get('target_shift_duration_hours', 8)
    target_roles = config.get('target_shift_roles', [])
    gmail_label = config.get('gmail_label', 'Deputy')
    notification_method = config.get('notification_method', 'console')
    
    print(f"Target shift duration: >={target_duration} hours")
    print(f"Target roles: {', '.join(target_roles)}")
    print(f"Gmail label: {gmail_label}")
    print("=" * 60)
    
    # Authenticate with Gmail
    print("\nAuthenticating with Gmail...")
    service = authenticate_gmail()
    if not service:
        print("Failed to authenticate with Gmail")
        return
    
    print("✓ Successfully authenticated")
    
    # Fetch Deputy emails
    print(f"\nFetching emails with label '{gmail_label}'...")
    messages = get_deputy_emails(service, gmail_label)
    print(f"✓ Found {len(messages)} email(s)")
    
    if not messages:
        return
    
    # Process each email
    print("\nProcessing emails...\n")
    matches_found = 0
    
    for i, message in enumerate(messages, 1):
        subject = get_email_subject(message)
        date = get_email_date(message)
        body = decode_email_body(message['payload'])
        
        print(f"[{i}/{len(messages)}] {subject[:50]}...")
        
        # Extract shift information
        shift_info = extract_shift_info(body, subject)
        
        if shift_info:
            # Check if it meets criteria
            meets_criteria, reason = check_criteria(
                shift_info,
                target_duration,
                target_roles
            )
            
            if meets_criteria:
                matches_found += 1
                notification_message = (
                    f"Email: {subject}\n"
                    f"Date: {date}\n"
                    f"Shift Role: {shift_info.get('role', 'N/A')}\n"
                    f"Duration: {shift_info.get('duration_hours', 'N/A')} hours\n"
                    f"Start Time: {shift_info.get('start_time', 'N/A')}\n"
                    f"End Time: {shift_info.get('end_time', 'N/A')}\n"
                    f"Match Reason: {reason}"
                )
                notify(notification_message, notification_method)
    
    # Summary
    print("\n" + "=" * 60)
    print(f"Processing complete!")
    print(f"Total emails processed: {len(messages)}")
    print(f"Matching shifts found: {matches_found}")
    print("=" * 60)


if __name__ == '__main__':
    main()
