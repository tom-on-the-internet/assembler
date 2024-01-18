package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	inputFile, outputFile, err := getReaderAndWriter()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = assemble(inputFile, outputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getReaderAndWriter() (*os.File, *os.File, error) {
	if len(os.Args) != 3 {
		return nil, nil, errors.New("usage: " + os.Args[0] + " input_file output_file")
	}

	inputFile, err := os.Open(os.Args[1])
	if err != nil {
		return nil, nil, err
	}

	outputFile, err := os.Create(os.Args[2])
	if err != nil {
		return nil, nil, err
	}

	return inputFile, outputFile, nil
}

func assemble(r io.Reader, w io.Writer) error {
	symbolMap := map[string]int{
		"R0":     0,
		"R1":     1,
		"R2":     2,
		"R3":     3,
		"R4":     4,
		"R5":     5,
		"R6":     6,
		"R7":     7,
		"R8":     8,
		"R9":     9,
		"R10":    10,
		"R11":    11,
		"R12":    12,
		"R13":    13,
		"R14":    14,
		"R15":    15,
		"SCREEN": 16384,
		"KBD":    24576,
		"SP":     0,
		"LCL":    1,
		"ARG":    2,
		"THIS":   3,
		"THAT":   4,
	}

	compSymbolMap := map[string]string{
		"0":   "0101010",
		"1":   "0111111",
		"-1":  "0111010",
		"D":   "0001100",
		"A":   "0110000",
		"!D":  "0001101",
		"!A":  "0110001",
		"-D":  "0001111",
		"-A":  "0110011",
		"D+1": "0011111",
		"A+1": "0110111",
		"D-1": "0001110",
		"A-1": "0110010",
		"D+A": "0000010",
		"D-A": "0010011",
		"A-D": "0000111",
		"D&A": "0000000",
		"D|A": "0010101",
		"M":   "1110000",
		"!M":  "1110001",
		"-M":  "1110011",
		"M+1": "1110111",
		"M-1": "1110010",
		"D+M": "1000010",
		"D-M": "1010011",
		"M-D": "1000111",
		"D&M": "1000000",
		"D|M": "1010101",
	}

	destMap := map[string]string{
		"M":   "001",
		"D":   "010",
		"MD":  "011",
		"A":   "100",
		"AM":  "101",
		"AD":  "110",
		"AMD": "111",
	}

	jumpMap := map[string]string{
		"JGT": "001",
		"JEQ": "010",
		"JGE": "011",
		"JLT": "100",
		"JNE": "101",
		"JLE": "110",
		"JMP": "111",
	}

	// first pass removes comments and fills the symbolMap
	// with the label values
	instructions := []string{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.Split(line, "//")[0]

		// ignore empty lines
		if len(line) == 0 {
			continue
		}

		// ignore comments
		if strings.HasPrefix(line, "//") {
			continue
		}

		// handle label
		if strings.HasPrefix(line, "(") {
			label := line[1:strings.Index(line, ")")]
			symbolMap[label] = len(instructions)

			continue
		}

		// at this point, the line is an instruction
		// this will appear in the binary output
		instructions = append(instructions, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	nextSymbolAddress := 16

	for _, instruction := range instructions {
		isAInstruction := strings.HasPrefix(instruction, "@")

		// a instructions
		if isAInstruction {
			address := instruction[1:]
			addressInt, err := strconv.Atoi(address)
			addressIsInMap := false

			if err != nil {
				// this means the address is a variable
				addressInt, addressIsInMap = symbolMap[address]
				if !addressIsInMap {
					addressInt = nextSymbolAddress

					// store for future look ups
					symbolMap[address] = addressInt

					nextSymbolAddress++
				}
			}

			binaryString := strconv.FormatInt(int64(addressInt), 2)
			// pad
			binaryString = fmt.Sprintf("%016s", binaryString)
			w.Write([]byte(binaryString + "\n"))

			continue
		}

		// c instructions

		// comp with a
		var compSymbol string
		if strings.Contains(instruction, "=") {
			compSymbol = strings.Split(instruction, "=")[1]
		} else {
			compSymbol = strings.Split(instruction, ";")[0]
		}

		comp := compSymbolMap[compSymbol]

		// dest
		dest := "000"

		d := strings.Split(instruction, "=")
		if len(d) == 2 {
			dest = destMap[d[0]]
		}

		// jmp
		jmp := "000"

		j := strings.Split(instruction, ";")
		if len(j) == 2 {
			jmp = jumpMap[j[1]]
		}

		w.Write([]byte("111" + comp + dest + jmp + "\n"))
	}

	return nil
}
