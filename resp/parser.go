package resp

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
		return "", fmt.Errorf("Expected simple string but received: %s", line)
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
		return "", fmt.Errorf("Expected error but received: %s", line)
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
		return "", fmt.Errorf("Expected integer but received: %s", line);
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
		return "", fmt.Errorf("Expected bulk string but received: %s", line);
	};

	stringLength, err := strconv.Atoi(line[1:]);
	if err != nil {
		fmt.Println("Error converting string length from string to integer");
		return "", err;
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
		return "", fmt.Errorf("Expected trailing CRLF after bulk string but received: %s", string(trail));
	}

	return string(buf), nil
}

func parseArrays(reader *bufio.Reader) ([]string, error) {
	// example of an array: *2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
	line, err := reader.ReadString('\n');
	if err != nil {
		fmt.Println("Error reading the input from the reader");
		return nil, err;
	};

	line = strings.TrimSpace(line);
	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("Expected array but received: %s", line)
	};

	arrayLength, err := strconv.Atoi(line[1:]);
	if err != nil {
		fmt.Println("Error converting array length from string to integer");
		return nil, fmt.Errorf("Invalid array length: %s", line[1:]);
	};

	result := make([]string, arrayLength);
	for i := 0; i < arrayLength; i++ {
		bulkString, err := parseBulkStrings(reader);
		if err != nil {
			fmt.Println("Error parsing bulk string in array");
			return nil, err;
		};

		result[i] = bulkString;
	}

	return result, nil;
};