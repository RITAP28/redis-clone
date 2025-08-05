package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func parseSimpleStrings(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n');
	if err != nil {
		fmt.Println("Error parsing simple string")
		return "Error parsing simple string", err
	};

	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] != '+' {
		// fmt.Printf("Expected simple string but received: %s", line)
		return "", fmt.Errorf("expected simple string but received: %s", line)
	}

	return line[1:], nil
}

func parseErrors(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n');
	if err != nil {
		fmt.Println(err);
		return "Error parsing", err;
	};

	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] != '-' {
		return "", fmt.Errorf("expected error but received: %s", line)
	}

	return line[1:], nil
}

func parseIntegers(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n');
	if err != nil {
		fmt.Println(err);
		return "", err;
	};

	line = strings.TrimSpace(line);
	if len(line) == 0 || line[0] != ':' {
		return "", fmt.Errorf("expected integer but received: %s", line);
	};

	return line[1:], nil;
}

func parseBulkStrings(reader *bufio.Reader) (string, error) {
	// example of a bulk string: $3\r\nfoo\r\n
	line, err := reader.ReadString('\n');
	if err != nil {
		fmt.Println("Error parsing bulk string");
		return "", err;
	};

	line = strings.TrimSpace(line);
	if len(line) == 0 || line[0] != '$' {
		return "", fmt.Errorf("expected bulk string but received: %s", line);
	};

	stringLength, err := strconv.Atoi(line[1:]);
	if err != nil {
		fmt.Println("Error converting string length from string to integer");
		return "", err;
	};

	// if the stringLength is -1, it means the bulk string is null
	// it will return nil
	if stringLength < 0 {
		fmt.Println("Bulk string length is negative, suggesting it is a null bulk string, returning nil");
		return "nil", nil;
	};

	buf := make([]byte, stringLength);
	_, err = io.ReadFull(reader, buf);
	if err != nil {
		fmt.Println("Error reading bulk string through io reader");
		return "", err;
	};

	trail := make([]byte, 2);
	_, err = io.ReadFull(reader, trail);
	if err != nil {
		fmt.Println("Error reading trailing bytes after bulk string");
		return "", err;
	};

	if string(trail) != "\r\n" {
		return "", fmt.Errorf("expected trailing CRLF after bulk string but received: %s", string(trail));
	}

	return string(buf), nil
}

func parseArrays(reader *bufio.Reader) ([]string, error) {
	// example of an array: *2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
	// the array can contain any data type, including bulk strings, integers, simple strings etc
	line, err := reader.ReadString('\n');
	if err != nil {
		fmt.Println("Error reading the input from the reader");
		return nil, err;
	};

	line = strings.TrimSpace(line);
	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("expected array but received: %s", line)
	};

	arrayLength, err := strconv.Atoi(line[1:]);
	// if the array length is -1, it means the array is empty
	if err != nil {
		fmt.Println("Error converting array length from string to integer");
		return nil, fmt.Errorf("invalid array length: %s", line[1:]);
	};

	// hanlding the case of null arrays
	if arrayLength < 0 {
		fmt.Println("Array length is negative, suggesting it is a null array, returning nil");
		return nil, nil;
	};

	result := make([]string, arrayLength);
	for i := range arrayLength {
		bulkString, err := parseBulkStrings(reader);
		if err != nil {
			fmt.Println("Error parsing bulk string in array");
			return nil, err;
		};

		result[i] = bulkString;
	}

	return result, nil;
};

func parseNulls(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n');
	if err != nil {
		fmt.Println("Error reading null value");
		return "", err;
	};

	line = strings.TrimSpace(line);
	if len(line) == 0 || line[0] != '_' {
		return "", fmt.Errorf("expected null value but received: %s", line);
	};

	return "", nil;
};

func parseBooleans(reader *bufio.Reader) (bool, error) {
	line, err := reader.ReadString('\n');
	if err != nil {
		fmt.Println("Error reading boolean value");
		return false, err;
	};

	line = strings.TrimSpace(line);
	if len(line) == 0 || line[0] != '#' {
		return false, fmt.Errorf("expected boolean value but received: %s", line);
	};

	if len(line) < 2 {
		return false, fmt.Errorf("incomplete boolean value: %s", line);
	}

	switch line[1] {
	case 't':
		return true, nil;
	case 'f':
		return true, nil;
	default:
		return false, fmt.Errorf("invalid boolean value: %s", line)
	}
}