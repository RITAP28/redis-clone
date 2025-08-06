package parser

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestHandleRESP_Arrays(t *testing.T) {
	testCases := []struct {
		name string
		input string
		expected []interface{}
		expectError bool
	}{
		{
			name: "Simple SET command",
			input: "*3\r\n$3\r\nSET\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expected: []interface{}{"SET", "hello", "world"},
			expectError: false,
		},
		{
			name: "Simple GET command",
			input: "*2\r\n$3\r\nGET\r\nhello\r\n",
			expected: []interface{}{"GET", "hello"},
			expectError: false,
		},
		{
			name: "Null Array",
			input: "*-1\r\n",
			expected: nil,
			expectError: false,
		},
		{
			name: "Empty Array",
			input: "*0\r\n",
			expected: []interface{}{},
			expectError: false,
		},
		{
			name: "Malformed Array (wrong count)",
			input: "*3\r\n$3\r\nGET\r\n$3\r\nhello\r\n",
			expectError: true,
		},
	};

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input));
			result, err := HandleRESP(reader);

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error, but got nil");
				}
				return // test passed
			};

			if err != nil {
				t.Fatalf("Did not expect an error, but got: %v\r\n", err);
			};

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %#v, but got %#v", tc.expected, result);
			};
		});
	};
};