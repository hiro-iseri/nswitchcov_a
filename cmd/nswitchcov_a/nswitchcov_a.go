package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"flag"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type TextType int
type State string
type Event string

const (
	StatusText TextType = iota
	EventText
)


func IncludePath(execPathSet [][]string, stateFlow []string) bool {
	if len(stateFlow) == 0 {
		return false
	}
	for _, execPath := range execPathSet {
		if len(execPath) == 0 {
			continue
		}
		for i := 0; i <= len(execPath) - len(stateFlow); i++ {
			if reflect.DeepEqual(execPath[i:i + len(stateFlow)], stateFlow) {
				fmt.Println("hit")
				return true
			}
		}
	}
	return false
}

/*
func main() {
	var (
		fpExePath = flag.String("exepath", "", "filepath of execution path list")
		fpStateFlow = flag.String("stateflow", "", "filepath of stateflow")
		encode = flag.String("encode", "", "encoding of input file")
	)
	flag.Parse()
	execPath := [][]string{}
	execPath = append(execPath, []string{"s1", "e1", "s2", "e2", "s3"})
	execPath = append(execPath, []string{"s5", "e2", "s2"})
	stateFlow := []string {"s2", "e2", "s3"}

	fmt.Println(*fpExePath, *fpStateFlow, *encode)

	fmt.Println(IncludePath(execPath, stateFlow))

}
*/


func pickupWord(word string) string {
	re, _ := regexp.Compile("^[\\s]+")
	trim_word := re.ReplaceAllString(word, "")
	re, _ = regexp.Compile("[\\s]+$")
	trim_word = re.ReplaceAllString(trim_word, "")
	return trim_word
}

func AddFlowPath(output [][]string, add_path []string) [][]string {
	for _, v := range output {
		if reflect.DeepEqual(v, add_path) {
			return output
		}
	}
	return append(output, add_path)
}

func ReadExecutionPath(fileName string) [][]string {

	fp, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	sjisScanner := bufio.NewScanner(transform.NewReader(fp, japanese.ShiftJIS.NewDecoder()))

	exePath := [][]string{}

	for sjisScanner.Scan() {
		tempExePath := []string{}
		targetText := sjisScanner.Text()
		currentType := StatusText
		word := ""
		if len(targetText) > 0 && targetText[0] == '#' {
			continue
		}
		for _, c := range targetText {
			if c == '-' {
				if currentType != StatusText {
					fmt.Println("error")
				}
				currentType = EventText
				tempExePath = append(tempExePath, pickupWord(word))
				continue
			}
			if c == '>' {
				if currentType != EventText {
					fmt.Println("error")
				}
				currentType = StatusText
				tempExePath = append(tempExePath, pickupWord(word))
				continue
			}
			word += string(c)
		}

		if currentType != StatusText {
			fmt.Println("error")
		}
		tempExePath = append(tempExePath, pickupWord(word))

		exePath = append(exePath, tempExePath)
	}

	fmt.Println(exePath)
	fmt.Println("*************")

	defer fp.Close()
	return exePath
}

// ReadStateFlow creates stateflow definition data from specified file
func ReadStateFlow(fileName string) (map[State]map[Event]State, error) {

	fp, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	sjisScanner := bufio.NewScanner(transform.NewReader(fp, japanese.ShiftJIS.NewDecoder()))

	state_map := make(map[State]map[Event]State)

	lineCount := 0
	for sjisScanner.Scan() {
		lineCount++
		if lineCount >= 200 {
			return nil, fmt.Errorf("File size limitation: maximum 200 lines")
		}
		currentState := ""
		current_event := ""
		targetText := sjisScanner.Text()
		fmt.Println(targetText)
		currentType := StatusText
		word := ""
		// Ignore blank line or line starting with #
		if len(targetText) > 0 && targetText[0] == '#' {
			continue
		}
		for _, c := range targetText {
			if c == '-' {
				if currentType != StatusText {
					return nil, fmt.Errorf("Error:Invalid Format in StateFlow File(line %d)", lineCount)
				}
				currentType = EventText
				trimmedWord := pickupWord(word)
				if len(trimmedWord) == 0 {
					return nil, fmt.Errorf("Error:Empty Keyword(line %d)", lineCount)
				}
				if currentState != "" {
					value, init := state_map[State(current_state)]
					if !init {
						value = make(map[Event]State)
					}
					value[Event(current_event)] = State(trimmedWord)
					state_map[State(currentState)] = value
				}
				currentState = trimmedWord
				word = ""
				continue
			}
			if c == '>' {
				if currentType != EventText {
					return nil, fmt.Errorf("Error:Invalid Format in StateFlow File(line %d)", lineCount)
				}
				currentType = StatusText
				current_event = pickupWord(word)
				if len(current_event) == 0 {
					return nil, fmt.Errorf("Error:Empty Keyword(line %d)", lineCount)
				}
				word = ""
				continue
			}
			word += string(c)
		}

		if currentType != StatusText {
			return nil, fmt.Errorf("Error:Invalid Format in StateFlow File(line %d)", lineCount)
		}
		trimmedWord := pickupWord(word)
		if currentState != "" {
			value, init := state_map[State(currentState)]
			if !init {
				value = make(map[Event]State)
			}
			value[Event(current_event)] = State(trimmedWord)
			state_map[State(currentState)] = value
		}
		currentState = pickupWord(word)
		word = ""

	}
	defer fp.Close()
	return state_map, nil
}

func main() {
	stateFlow, err := ReadStateFlow("input.txt")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	stateFlowPath := CreateList(stateFlow)

	//execPath := ReadExecutionPath("exe.txt")
	//fmt.Println(execPath)
	fmt.Println(stateFlowPath)

}

func CreateList(m map[State]map[Event]State) [][]string {
	return CreateNSwitchPathSet(m, 2)
}

// CreateNSwitchPathSet creates path set covering N-switch coverage
func CreateNSwitchPathSet(m map[State]map[Event]State, nValue int) [][]string {
	var stackRecPath [][]string
	var outputs [][]string
	recCount := 0

	for k := range m {
		fmt.Println("start0")
		stackRecPath = [][]string{}
		stackRecPath = append(stackRecPath, []string{})
		stackRecPath[0] = append(stackRecPath[0], string(k))

		createNSwitchPathSetRec(&stackRecPath, &outputs, m, nValue+1, &recCount, k)
	}
	fmt.Println(outputs)
	return outputs
}

// pickup recursively path set covering N-switch coverage
func createNSwitchPathSetRec(stackRecPath *[][]string, outputs *[][]string, m map[State]map[Event]State, recLimit int, recCount *int, nextState State) {
	*recCount++

	for event, targetState := range m[nextState] {
		fmt.Println("start", *recCount, *stackRecPath)
		if len(*stackRecPath) < *recCount+1 {
			*stackRecPath = append(*stackRecPath, []string{})
		}
		(*stackRecPath)[*recCount] = (*stackRecPath)[*recCount-1]
		(*stackRecPath)[*recCount] = append((*stackRecPath)[*recCount], string(event))
		(*stackRecPath)[*recCount] = append((*stackRecPath)[*recCount], string(targetState))

		if recLimit <= *recCount {
			fmt.Println((*stackRecPath)[*recCount])
			*outputs = AddFlowPath(*outputs, (*stackRecPath)[*recCount])
			(*stackRecPath)[*recCount] = (*stackRecPath)[*recCount-1]
			continue
		}

		if len(m[targetState]) == 0 {
			(*stackRecPath)[*recCount] = (*stackRecPath)[*recCount-1]
		}

		createNSwitchPathSetRec(stackRecPath, outputs, m, recLimit, recCount, targetState)
	}
	*recCount--
}
