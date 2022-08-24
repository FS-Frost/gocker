package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type container struct {
	ID    string
	Names string
}

func main() {
	binary, err := exec.LookPath("docker")
	checkError(err, "error looking for docker binary path")

	containers, err := getContainers(binary)
	checkError(err, "error getting containers")
	if containers == nil {
		fmt.Println("No containers running.")
		return
	}

	if len(containers) == 0 {
		fmt.Println("No available containers found")
		return
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Names < containers[j].Names
	})

	selectedContainerIndex, err := getSelectedContainerIndex(containers)
	checkError(err, "error getting selected container")

	selectedContainer := containers[selectedContainerIndex]
	fmt.Printf("Container: %s (%s)\n", selectedContainer.Names, selectedContainer.ID)

	fmt.Println()
	commands, err := getCommands()
	checkError(err, "error getting selected command")
	fmt.Printf("Command: %s\n", strings.Join(commands, " "))

	fmt.Println()
	execCommandsOnContainer(binary, commands, selectedContainer.Names)
}

func checkError(err error, format string, a ...any) {
	if err != nil {
		s := fmt.Sprintf(format, a...)
		s = fmt.Sprintf("%s: %v", s, err)
		panic(s)
	}
}

func getContainers(binary string) ([]container, error) {
	cmd := exec.Command(binary, "ps", "--format", "{{json .}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error reading containers info: %v", err)
	}

	stringOut := string(out)
	lines := strings.Split(stringOut, "\n")
	if len(lines) < 1 {
		return nil, nil
	}

	containers := []container{}
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		if l == "" {
			continue
		}

		bs := []byte(l)
		c := &container{}
		err = json.Unmarshal(bs, c)
		if err != nil {
			return nil, fmt.Errorf("error parsing docker ps output at line %d: %v\nJSON: <%s>", i+1, err, l)
		}

		containers = append(containers, *c)
	}

	return containers, nil
}

func getInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	trimmedText := strings.ReplaceAll(text, "\n", "")
	return trimmedText, err
}

func getSelectedContainerIndex(containers []container) (int, error) {
	fmt.Println("Containers:")
	for i, c := range containers {
		fmt.Printf("%d. %s\n", i+1, c.Names)
	}

	for {
		fmt.Printf("\nEnter container number or name: ")
		stringSelectedNumber, err := getInput()
		if err != nil {
			return 0, fmt.Errorf("error reading selection from input: %v", err)
		}

		userWroteListCommand, selectedIndex := containerSliceContainsName(containers, stringSelectedNumber)
		if userWroteListCommand {
			return selectedIndex, nil
		}

		selectedNumber, err := strconv.Atoi(stringSelectedNumber)
		if err != nil {
			fmt.Printf("error parsing selected number '%s'\n", stringSelectedNumber)
			continue
		}

		selectedIndex = selectedNumber - 1
		if selectedIndex >= 0 && selectedIndex < len(containers) {
			return selectedIndex, nil
		}

		fmt.Printf("Invalid selection '%d', try again.\n", selectedNumber)
	}
}

func getCommands() ([]string, error) {
	commands := []string{"bash", "sh", "other"}
	fmt.Println("Commands:")
	for i, c := range commands {
		fmt.Printf("%d. %s\n", i+1, c)
	}

	var command string
	for {
		fmt.Printf("\nEnter command number or raw command to execute: ")
		stringSelectedNumber, err := getInput()
		if err != nil {
			return nil, fmt.Errorf("error reading selection from input: %v", err)
		}

		userWroteListCommand, _ := stringSliceContains(commands, stringSelectedNumber)
		if userWroteListCommand {
			command = stringSelectedNumber
			break
		}

		selectedNumber, err := strconv.Atoi(stringSelectedNumber)
		if err != nil {
			fmt.Printf("error parsing selected number '%s'\n", stringSelectedNumber)
			continue
		}

		selectedIndex := selectedNumber - 1
		if selectedIndex >= 0 && selectedIndex < len(commands) {
			command = commands[selectedIndex]
			break
		}

		fmt.Printf("Invalid selection '%d', try again.\n", selectedNumber)
	}

	if command != "other" {
		return []string{command}, nil
	}

	fmt.Printf("Enter command: ")
	rawInputCommands, err := getInput()
	if err != nil {
		return nil, fmt.Errorf("error reading command: %v", err)
	}

	inputCommands := strings.Split(rawInputCommands, " ")
	return inputCommands, nil
}

func stringSliceContains(slice []string, s string) (bool, int) {
	for i, v := range slice {
		if v == s {
			return true, i
		}
	}

	return false, 0
}

func containerSliceContainsName(slice []container, s string) (bool, int) {
	for i, c := range slice {
		if c.Names == s {
			return true, i
		}
	}

	return false, 0
}

func execCommandsOnContainer(binary string, commands []string, container string) {
	args := []string{"exec", "-it", container}
	args = append(args, commands...)
	cmd := exec.Command(binary, args...)
	fmt.Println(cmd.String())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error while in container: %v\n", err)
	}

	exitCode := cmd.ProcessState.ExitCode()
	fmt.Printf("Exited from %s with code %d\n", container, exitCode)
}
