package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"

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

func includePath(execPathSet [][]string, stateFlow []string) bool {
	if len(stateFlow) == 0 {
		return false
	}
	for _, execPath := range execPathSet {
		if len(execPath) == 0 {
			continue
		}
		for i := 0; i <= len(execPath)-len(stateFlow); i++ {
			if reflect.DeepEqual(execPath[i:i+len(stateFlow)], stateFlow) {
				return true
			}
		}
	}
	return false
}

func main() {
	var (
		fpExePath   = flag.String("exepath", "", "filepath of execution path list")
		fpStateFlow = flag.String("stateflow", "", "filepath of stateflow")
		//charcode    = flag.String("charcode", "", "encoding of input file")
	)
	flag.Parse()

	execPath, err := ReadExecutionPath(*fpExePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	stateFlowPath, err := ReadExecutionPath(*fpStateFlow)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	stateFlowMap, _ := CreateStateFlowMap(stateFlowPath)

	stateFlowPathSet := CreateNSwitchPathSet(stateFlowMap, 2)

	fmt.Println("*******map*******")
	fmt.Println(stateFlowMap)
	fmt.Println("*******stateflow path*******")
	fmt.Println(stateFlowPathSet)
	fmt.Println("*******exec path*******")
	fmt.Println(execPath)

	sumNSwitchPath := len(stateFlowPathSet)
	lenExecPath := len(execPath)

	fmt.Printf("number of execution path:%d\n", lenExecPath)
	fmt.Printf("number of n-switch path:%d\n", sumNSwitchPath)

	sumCoveringPath := 0

	for _, path := range stateFlowPathSet {
		if includePath(execPath, path) {
			sumCoveringPath++
		}
	}

	var coverage float64
	coverage = float64(sumCoveringPath) / float64(sumNSwitchPath) * 100.0
	fmt.Printf("n-switch coverage:%.2f%%(%d/%d)\n", coverage, sumCoveringPath, sumNSwitchPath)
}

func pickupWord(word string) string {
	re, _ := regexp.Compile("^[\\s]+")
	trimWord := re.ReplaceAllString(word, "")
	re, _ = regexp.Compile("[\\s]+$")
	trimWord = re.ReplaceAllString(trimWord, "")
	return trimWord
}

func addFlowPath(output [][]string, addPath []string) [][]string {
	for _, v := range output {
		if reflect.DeepEqual(v, addPath) {
			return output
		}
	}
	return append(output, addPath)
}

func ReadExecutionPath(fileName string) ([][]string, error) {
	fp, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	sjisScanner := bufio.NewScanner(transform.NewReader(fp, japanese.ShiftJIS.NewDecoder()))

	exePath := [][]string{}

	lineCount := 0
	for sjisScanner.Scan() {
		lineCount++
		if lineCount >= 200 {
			return nil, fmt.Errorf("File size limitation: maximum 200 lines:%s", fileName)
		}
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
					return nil, fmt.Errorf("Error:Invalid Format in File(%s line %d)", fileName, lineCount)
				}
				currentType = EventText
				trimmedWord := pickupWord(word)
				if len(trimmedWord) == 0 {
					return nil, fmt.Errorf("Error:Empty Keyword(%s line %d)", fileName, lineCount)
				}
				tempExePath = append(tempExePath, pickupWord(trimmedWord))
				word = ""
				continue
			}
			if c == '>' {
				if currentType != EventText {
					return nil, fmt.Errorf("Error:Invalid Format in File(%s line %d)", fileName, lineCount)
				}
				currentType = StatusText
				tempExePath = append(tempExePath, pickupWord(word))
				word = ""
				continue
			}
			word += string(c)
		}

		if currentType != StatusText {
			return nil, fmt.Errorf("Error:Invalid Format in StateFlow File(%s line %d)", fileName, lineCount)
		}
		tempExePath = append(tempExePath, pickupWord(word))

		exePath = append(exePath, tempExePath)
	}

	fmt.Println(exePath)
	defer fp.Close()
	return exePath, nil
}

// CreateStateFlowMap creates stateflow definition data from specified file
func CreateStateFlowMap(flowpath [][]string) (map[State]map[Event]State, error) {
	fmt.Println("************************************")
	fmt.Println(flowpath)
	fmt.Println("@@@")
	stateMap := make(map[State]map[Event]State)

	for _, targetText := range flowpath {
		currentState := ""
		currentEvent := ""
		currentType := StatusText

		for _, word := range targetText {
			if currentType == StatusText {
				currentType = EventText

				if currentEvent != "" {
					value, init := stateMap[State(currentState)]
					if !init {
						value = make(map[Event]State)
					}
					value[Event(currentEvent)] = State(word)
					stateMap[State(currentState)] = value
					currentEvent = ""
				}
				currentState = word
			} else if currentType == EventText {
				currentType = StatusText
				currentEvent = word
				continue
			}
		}
	}
	fmt.Println("************************************")

	return stateMap, nil
}

// ReadStateFlow creates stateflow definition data from specified file
func ReadStateFlow(fileName string) (map[State]map[Event]State, error) {
	fp, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	sjisScanner := bufio.NewScanner(transform.NewReader(fp, japanese.ShiftJIS.NewDecoder()))

	stateMap := make(map[State]map[Event]State)

	lineCount := 0
	for sjisScanner.Scan() {
		lineCount++
		if lineCount >= 200 {
			return nil, fmt.Errorf("File size limitation: maximum 200 lines:%s", fileName)
		}
		currentState := ""
		currentEvent := ""
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
					return nil, fmt.Errorf("Error:Invalid Format in File(%s line %d)", fileName, lineCount)
				}
				currentType = EventText
				trimmedWord := pickupWord(word)
				if len(trimmedWord) == 0 {
					return nil, fmt.Errorf("Error:Empty Keyword(%s line %d)", fileName, lineCount)
				}
				if currentState != "" {
					value, init := stateMap[State(currentState)]
					if !init {
						value = make(map[Event]State)
					}
					value[Event(currentEvent)] = State(trimmedWord)
					stateMap[State(currentState)] = value
				}
				currentState = trimmedWord
				word = ""
				continue
			}
			if c == '>' {
				if currentType != EventText {
					return nil, fmt.Errorf("Error:Invalid Format in File(%s line %d)", fileName, lineCount)
				}
				currentType = StatusText
				currentEvent = pickupWord(word)
				if len(currentEvent) == 0 {
					return nil, fmt.Errorf("Error:Empty Keyword(%s line %d)", fileName, lineCount)
				}
				word = ""
				continue
			}
			word += string(c)
		}

		if currentType != StatusText {
			return nil, fmt.Errorf("Error:Invalid Format in StateFlow File(%s line %d)", fileName, lineCount)
		}
		trimmedWord := pickupWord(word)
		if currentState != "" {
			value, init := stateMap[State(currentState)]
			if !init {
				value = make(map[Event]State)
			}
			value[Event(currentEvent)] = State(trimmedWord)
			stateMap[State(currentState)] = value
		}
		currentState = pickupWord(word)
		word = ""

	}
	defer fp.Close()
	return stateMap, nil
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
			*outputs = addFlowPath(*outputs, (*stackRecPath)[*recCount])
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
