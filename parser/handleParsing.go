package parser

import (
	"bufio"
	"fmt"
)

func HandleRESP(reader *bufio.Reader) (interface{}, error) {
	// reads the first byte of the incoming request
	typeIndicator, err := reader.ReadByte();
	if err != nil {
		fmt.Println("Error reading the incoming request");
		return nil, err
	};

	// analysing the first byte and then calling the corresponding function
	switch typeIndicator {
	case '+':
		return parseSimpleStrings(reader);
	case '-':
		return parseErrors(reader);
	case ':':
		return parseIntegers(reader);
	case '$':
		return parseBulkStrings(reader);
	case '*':
		return parseArrays(reader);
	case '_':
		return parseNulls(reader);
	case '#':
		return parseBooleans(reader);
	default:
		return nil, fmt.Errorf("unknown RESP type indicator: %c", typeIndicator);
	}
}