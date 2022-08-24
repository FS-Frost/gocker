package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
)

type config struct {
	Containers map[string][]string `json:"containers"`
}

func newConfig() *config {
	conf := &config{
		Containers: make(map[string][]string),
	}
	return conf
}

type container struct {
	ID    string
	Names string
}

func main() {
	conf, err := getConfig()
	if err != nil {
		fmt.Printf("config error: %v\n", err)
	}

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

	selectedContainerIndex, err := getSelectedContainerIndex(conf, containers)
	checkError(err, "error getting selected container")

	selectedContainer := containers[selectedContainerIndex]
	fmt.Printf("Container: %s (%s)\n", selectedContainer.Names, selectedContainer.ID)

	fmt.Println()
	defaultCommands := conf.Containers[selectedContainer.Names]
	commands, err := getCommands(defaultCommands...)
	checkError(err, "error getting selected command")
	fmt.Printf("Command: %s\n", strings.Join(commands, " "))

	conf.Containers[selectedContainer.Names] = commands
	err = saveConfig(conf)
	if err != nil {
		fmt.Printf("error saving config: %v\n", err)
	}

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

func getSelectedContainerIndex(conf *config, containers []container) (int, error) {
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

func getCommands(defaultCommands ...string) ([]string, error) {
	defaultCommandItem := ""
	if len(defaultCommands) > 0 {
		defaultCommandItem = strings.Join(defaultCommands, " ")
	}

	commands := []string{"bash", "sh", "other"}
	fmt.Println("Commands:")

	if defaultCommandItem != "" {
		fmt.Printf("Press ENTER to select default (%s)\n", defaultCommandItem)
	}

	for i, c := range commands {
		defaultMsg := ""
		if c == defaultCommandItem {
			defaultMsg = " (default)"
		}

		fmt.Printf("%d. %s%s\n", i+1, c, defaultMsg)
	}

	var command string
	for {
		fmt.Printf("\nEnter command number: ")
		stringSelectedNumber, err := getInput()
		if err != nil {
			return nil, fmt.Errorf("error reading selection from input: %v", err)
		}

		userProvidedListCommand, _ := stringSliceContains(commands, stringSelectedNumber)
		if userProvidedListCommand {
			command = stringSelectedNumber
			break
		}

		if stringSelectedNumber == "" && defaultCommandItem != "" {
			return defaultCommands, nil
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

	fmt.Printf("Enter raw command: ")
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
	if err != nil && !strings.HasPrefix(err.Error(), "exit status") {
		fmt.Printf("Container error: %v\n", err)
	}

	exitCode := cmd.ProcessState.ExitCode()
	fmt.Printf("Exited from %s with code %d\n", container, exitCode)
}

func getConfigPath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %v", err)
	}

	confDir := path.Join(dir, ".gocker")
	err = os.MkdirAll(confDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error creating config directory: %v", err)
	}

	filePath := path.Join(confDir, "config.json")
	return filePath, nil
}

func getConfig() (*config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("error getting config path: %v", err)
	}

	bs, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return newConfig(), nil
	}

	if err != nil {
		return newConfig(), fmt.Errorf("error reading config file: %v", err)
	}

	conf := &config{}
	err = json.Unmarshal(bs, conf)
	if err != nil {
		return newConfig(), fmt.Errorf("error parsing config %s: %v", configPath, err)
	}

	return conf, nil
}

func saveConfig(conf *config) error {
	if conf == nil {
		return fmt.Errorf("nil config provided")
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %v", err)
	}

	confDir := path.Join(dir, ".gocker")
	err = os.MkdirAll(confDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	filePath := path.Join(confDir, "config.json")
	bs, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return fmt.Errorf("error encoding to json: %v", err)
	}

	err = os.WriteFile(filePath, bs, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writting config: %v", err)
	}

	return nil
}
