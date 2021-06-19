package main

import (
	"testing"
)

func doTestTelegramIsCommand(t *testing.T, message string, command string, expected bool) {
	if telegramIsCommand(message, command) != expected {
		t.Errorf("telegramIsCommand(\"%s\", \"%s\") unexpected result", message, command)
	}
}

func TestTelegramIsCommand(t *testing.T) {
	setting.telegramBotName = "test_bot"

	// Recognize command
	doTestTelegramIsCommand(t, "/trace", "trace", true)
	doTestTelegramIsCommand(t, "/trace", "trace1234", false)
	doTestTelegramIsCommand(t, "/trace", "tra", false)
	doTestTelegramIsCommand(t, "/trace", "abcdefg", false)

	// Recognize command with parameters
	doTestTelegramIsCommand(t, "/trace google.com", "trace", true)
	doTestTelegramIsCommand(t, "/trace google.com", "trace1234", false)
	doTestTelegramIsCommand(t, "/trace google.com", "tra", false)
	doTestTelegramIsCommand(t, "/trace google.com", "abcdefg", false)

	// Recognize command with bot name
	doTestTelegramIsCommand(t, "/trace@test_bot", "trace", true)
	doTestTelegramIsCommand(t, "/trace@test_bot", "trace1234", false)
	doTestTelegramIsCommand(t, "/trace@test_bot", "tra", false)
	doTestTelegramIsCommand(t, "/trace@test_bot", "abcdefg", false)
	doTestTelegramIsCommand(t, "/trace@test_bot_123", "trace", false)
	doTestTelegramIsCommand(t, "/trace@test_", "trace", false)

	// Recognize command with bot name and parameters
	doTestTelegramIsCommand(t, "/trace@test_bot google.com", "trace", true)
	doTestTelegramIsCommand(t, "/trace@test_bot google.com", "trace1234", false)
	doTestTelegramIsCommand(t, "/trace@test_bot google.com", "tra", false)
	doTestTelegramIsCommand(t, "/trace@test_bot google.com", "abcdefg", false)
	doTestTelegramIsCommand(t, "/trace@test_bot_123 google.com", "trace", false)
	doTestTelegramIsCommand(t, "/trace@test google.com", "trace", false)
}
