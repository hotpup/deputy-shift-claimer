#!/usr/bin/env python3
"""
Unit tests for Deputy Shift Claimer
"""

import unittest
from deputy_shift_claimer import (
    extract_shift_info,
    calculate_duration,
    check_criteria
)


class TestShiftInfoExtraction(unittest.TestCase):
    """Test shift information extraction from email content."""
    
    def test_extract_role_from_body(self):
        """Test extracting role from email body."""
        email_body = """
        Hello,
        
        A shift is available!
        Shift: Bartender
        Date: January 5, 2026
        """
        subject = "Deputy Shift Available"
        
        result = extract_shift_info(email_body, subject)
        self.assertIsNotNone(result)
        self.assertEqual(result['role'], 'Bartender')
    
    def test_extract_time_range(self):
        """Test extracting time range from email."""
        email_body = """
        Shift Details:
        Position: Server
        Time: 9:00 AM - 5:00 PM
        """
        subject = "New Shift"
        
        result = extract_shift_info(email_body, subject)
        self.assertIsNotNone(result)
        self.assertEqual(result['start_time'], '9:00 AM')
        self.assertEqual(result['end_time'], '5:00 PM')
        self.assertEqual(result['duration_hours'], 8.0)
    
    def test_extract_explicit_duration(self):
        """Test extracting explicit duration."""
        email_body = """
        Available shift
        Role: Manager
        Duration: 10 hours
        """
        subject = "Shift Available"
        
        result = extract_shift_info(email_body, subject)
        self.assertIsNotNone(result)
        self.assertEqual(result['duration_hours'], 10.0)
    
    def test_extract_role_from_subject(self):
        """Test extracting role from subject when not in body."""
        email_body = "A shift is available for pickup."
        subject = "Shift: Server - Available Now"
        
        result = extract_shift_info(email_body, subject)
        self.assertIsNotNone(result)
        self.assertEqual(result['role'], 'Server')
    
    def test_no_shift_info(self):
        """Test when no shift info is found."""
        email_body = "This is a general email with no shift information."
        subject = "General Update"
        
        result = extract_shift_info(email_body, subject)
        # Should return None or empty dict
        if result:
            self.assertIsNone(result['role'])
            self.assertIsNone(result['duration_hours'])


class TestDurationCalculation(unittest.TestCase):
    """Test duration calculation between times."""
    
    def test_standard_shift(self):
        """Test standard 8-hour shift."""
        duration = calculate_duration("9:00 AM", "5:00 PM")
        self.assertEqual(duration, 8.0)
    
    def test_overnight_shift(self):
        """Test shift that goes past midnight."""
        duration = calculate_duration("10:00 PM", "6:00 AM")
        self.assertEqual(duration, 8.0)
    
    def test_short_shift(self):
        """Test short shift."""
        duration = calculate_duration("2:00 PM", "6:00 PM")
        self.assertEqual(duration, 4.0)


class TestCriteriaChecking(unittest.TestCase):
    """Test shift criteria checking."""
    
    def test_duration_match(self):
        """Test matching duration criteria."""
        shift_info = {
            'role': 'Server',
            'duration_hours': 8.0,
            'start_time': '9:00 AM',
            'end_time': '5:00 PM'
        }
        
        meets, reason = check_criteria(shift_info, 8.0, ['Bartender', 'Server'])
        self.assertTrue(meets)
        self.assertIn('Duration', reason)
        self.assertIn('Role', reason)
    
    def test_role_match_only(self):
        """Test matching only role criteria."""
        shift_info = {
            'role': 'Bartender',
            'duration_hours': 4.0
        }
        
        meets, reason = check_criteria(shift_info, 8.0, ['Bartender'])
        self.assertTrue(meets)
        self.assertIn('Bartender', reason)
    
    def test_duration_match_only(self):
        """Test matching only duration criteria."""
        shift_info = {
            'role': 'Cook',
            'duration_hours': 10.0
        }
        
        meets, reason = check_criteria(shift_info, 8.0, ['Bartender', 'Server'])
        self.assertTrue(meets)
        self.assertIn('Duration', reason)
    
    def test_no_match(self):
        """Test when no criteria match."""
        shift_info = {
            'role': 'Cook',
            'duration_hours': 4.0
        }
        
        meets, reason = check_criteria(shift_info, 8.0, ['Bartender', 'Server'])
        self.assertFalse(meets)
        self.assertEqual(reason, "")
    
    def test_partial_role_match(self):
        """Test partial role name matching."""
        shift_info = {
            'role': 'Head Bartender',
            'duration_hours': 6.0
        }
        
        meets, reason = check_criteria(shift_info, 8.0, ['Bartender'])
        self.assertTrue(meets)
        self.assertIn('Bartender', reason)


if __name__ == '__main__':
    unittest.main()
