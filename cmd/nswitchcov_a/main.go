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

type textType int

// State of Stateflow
type State string

// Event of Stateflow
type Event string

var TargetEnv = "windows"
var NSwitchCovAVersion = "1.00"

const (
	statusText textType = iota
	eventText
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

func pickupWord(word string) string {
	re, _ := regexp.Compile("^[\\s]+")
	trimWord := re.ReplaceAllString(word, "")
	re, _ = regexp.Compile("[\\s]+$")
	trimWord = re.ReplaceAllString(trimWord, "")
	return trimWord
}

func addFlowPath(output [][]string, addPath []string) [][]string {
	addCopyPath := make([]string, len(addPath))
	copy(addCopyPath, addPath)

	for _, v := range output {
		if reflect.DeepEqual(v, addPath) {
			return output
		}
	}
	return append(output, addCopyPath)
}

// ReadExecutionPath creates stateflow path from specified file
func ReadExecutionPath(fileName string, encode string) ([][]string, error) {
	fp, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	var scanner *bufio.Scanner
	if encode == "utf8" {
		scanner = bufio.NewScanner(fp)
	} else if encode == "euc" {
		scanner = bufio.NewScanner(transform.NewReader(fp, japanese.EUCJP.NewDecoder()))
	} else {
		scanner = bufio.NewScanner(transform.NewReader(fp, japanese.ShiftJIS.NewDecoder()))
	}

	exePath := [][]string{}

	lineCount := 0
	for scanner.Scan() {
		lineCount++
		if lineCount >= 200 {
			return nil, fmt.Errorf("File size limitation: maximum 200 lines:%s", fileName)
		}
		tempExePath := []string{}
		targetText := scanner.Text()
		currentType := statusText
		word := ""
		if len(targetText) > 0 && targetText[0] == '#' {
			continue
		}
		for _, c := range targetText {
			if c == '-' {
				if currentType != statusText {
					return nil, fmt.Errorf("Error:Invalid Format in File(%s line %d)", fileName, lineCount)
				}
				currentType = eventText
				trimmedWord := pickupWord(word)
				if len(trimmedWord) == 0 {
					return nil, fmt.Errorf("Error:Empty Keyword(%s line %d)", fileName, lineCount)
				}
				tempExePath = append(tempExePath, pickupWord(trimmedWord))
				word = ""
				continue
			}
			if c == '>' {
				if currentType != eventText {
					return nil, fmt.Errorf("Error:Invalid Format in File(%s line %d)", fileName, lineCount)
				}
				currentType = statusText
				tempExePath = append(tempExePath, pickupWord(word))
				word = ""
				continue
			}
			word += string(c)
		}

		if currentType != statusText {
			return nil, fmt.Errorf("Error:Invalid Format in File(%s line %d)", fileName, lineCount)
		}
		tempExePath = append(tempExePath, pickupWord(word))

		exePath = append(exePath, tempExePath)
	}

	defer fp.Close()
	return exePath, nil
}

// CreateStateFlowMap creates stateflow definition data from path data
func CreateStateFlowMap(flowpath [][]string) (map[State]map[Event]State, error) {
	stateMap := make(map[State]map[Event]State)

	for _, targetText := range flowpath {
		currentState := ""
		currentEvent := ""
		currentType := statusText

		for _, word := range targetText {
			if currentType == statusText {
				currentType = eventText

				if currentEvent != "" {
					value, init := stateMap[State(currentState)]
					if !init {
						value = make(map[Event]State)
					}
					state, init := value[Event(currentEvent)]
					if init {
						if state != State(word) {
							return nil, fmt.Errorf("Error:There are different post-transition states in pre-transition:%s & event:%s", word, currentEvent)
						}
					}
					value[Event(currentEvent)] = State(word)
					stateMap[State(currentState)] = value
					currentEvent = ""
				}
				currentState = word
			} else if currentType == eventText {
				currentType = statusText
				currentEvent = word
				continue
			}
		}
	}
	return stateMap, nil
}

// CreateNSwitchPathSet creates path set covering N-switch coverage
func CreateNSwitchPathSet(m map[State]map[Event]State, nValue int) [][]string {
	var stackRecPath [][]string
	var outputs [][]string
	recCount := 0

	for k := range m {
		stackRecPath = [][]string{}
		stackRecPath = append(stackRecPath, []string{})
		stackRecPath[0] = append(stackRecPath[0], string(k))

		createNSwitchPathSetRec(&stackRecPath, &outputs, m, nValue+1, &recCount, k)
	}
	return outputs
}

// pickup recursively path set covering N-switch coverage
func createNSwitchPathSetRec(stackRecPath *[][]string, outputs *[][]string, m map[State]map[Event]State, recLimit int, recCount *int, nextState State) {
	*recCount++

	for event, targetState := range m[nextState] {
		if len(*stackRecPath) < *recCount+1 {
			*stackRecPath = append(*stackRecPath, []string{})
		}
		(*stackRecPath)[*recCount] = (*stackRecPath)[*recCount-1]
		(*stackRecPath)[*recCount] = append((*stackRecPath)[*recCount], string(event))
		(*stackRecPath)[*recCount] = append((*stackRecPath)[*recCount], string(targetState))

		if recLimit <= *recCount {
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

func nSwitchCovAMain(fpExePath string, fpStateFlow string, n int, encode string) {
	execPath, err := ReadExecutionPath(fpExePath, encode)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	stateFlowPath, err := ReadExecutionPath(fpStateFlow, encode)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	stateFlowMap, err := CreateStateFlowMap(stateFlowPath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	stateFlowPathSet := CreateNSwitchPathSet(stateFlowMap, n)

	sumNSwitchPath := len(stateFlowPathSet)
	lenExecPath := len(execPath)

	fmt.Printf("number of execution path:%d\n", lenExecPath)
	fmt.Printf("number of n-switch path:%d\n", sumNSwitchPath)

	sumCoveringPath := 0

	fmt.Println(stateFlowPathSet)

	for _, path := range stateFlowPathSet {
		if includePath(execPath, path) {
			sumCoveringPath++
		}
	}

	var coverage float64
	if sumNSwitchPath != 0 {
		coverage = float64(sumCoveringPath) / float64(sumNSwitchPath) * 100.0
	} else {
		coverage = 0
	}
	fmt.Printf("n-switch coverage:%.2f%%(%d/%d)\n", coverage, sumCoveringPath, sumNSwitchPath)
}

func main() {
	fmt.Printf("nswitchcov_a(ver:%s)\n", NSwitchCovAVersion)
	var (
		fpExePath   = flag.String("exepath", "", "filepath of execution path list")
		fpStateFlow = flag.String("stateflow", "", "filepath of stateflow")
		nOfSwitch   = flag.Int("n", -1, "n of n-switch coverage")
		charcode    = flag.String("charcode", "", "encoding of input file")
	)
	flag.Parse()

	if *fpExePath == "" {
		fmt.Println("Error:--exepath is not specified")
		os.Exit(1)
	}
	if *fpStateFlow == "" {
		fmt.Println("Error:--stateflow is not specified")
		os.Exit(1)
	}
	if *nOfSwitch < 0 {
		fmt.Println("Error:--n（n >= 0）is not specified")
		os.Exit(1)
	}
	if *charcode == "" {
		if TargetEnv == "darwin" {
			*charcode = "utf8"
		} else {
			*charcode = "utf8"
		}
	}

	nSwitchCovAMain(*fpExePath, *fpStateFlow, *nOfSwitch, *charcode)
}
