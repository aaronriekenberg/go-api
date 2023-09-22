package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/sync/semaphore"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/utils"
)

func NewAllCommandsHandler(commandConfiguration config.CommandConfiguration) http.Handler {
	jsonBuffer, err := json.Marshal(commandConfiguration.Commands)
	if err != nil {
		slog.Error("NewAllCommandsHandler json.Marshal error",
			"error", err)
		os.Exit(1)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(utils.ContentTypeHeaderKey, utils.ContentTypeApplicationJSON)
		io.Copy(w, bytes.NewReader(jsonBuffer))
	})
}

type commandInfoDTO struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	Args        []string `json:"args"`
}

func commandInfoToDTO(commandInfo config.CommandInfo) commandInfoDTO {
	return commandInfoDTO{
		ID:          commandInfo.ID,
		Description: commandInfo.Description,
		Command:     commandInfo.Command,
		Args:        commandInfo.Args,
	}
}

type runCommandsHandler struct {
	commandSemaphore        *semaphore.Weighted
	requestTimeout          time.Duration
	semaphoreAcquireTimeout time.Duration
	idToCommandInfo         map[string]commandInfoDTO
}

func NewRunCommandsHandler(commandConfiguration config.CommandConfiguration) http.Handler {
	requestTimeout, err := time.ParseDuration(commandConfiguration.RequestTimeoutDuration)
	if err != nil {
		slog.Error("error parsing RequestTimeoutDuration",
			"RequestTimeoutDuration", commandConfiguration.RequestTimeoutDuration,
			"error", err)
		os.Exit(1)
	}

	semaphoreAcquireTimeout, err := time.ParseDuration(commandConfiguration.SemaphoreAcquireTimeoutDuration)
	if err != nil {
		slog.Error("error parsing SemaphoreAcquireTimeoutDuration",
			"SemaphoreAcquireTimeoutDuration", commandConfiguration.SemaphoreAcquireTimeoutDuration,
			"error", err)
		os.Exit(1)
	}

	idToCommandInfo := make(map[string]commandInfoDTO)
	for _, commandInfo := range commandConfiguration.Commands {
		idToCommandInfo[commandInfo.ID] = commandInfoToDTO(commandInfo)
	}

	return &runCommandsHandler{
		commandSemaphore:        semaphore.NewWeighted(commandConfiguration.MaxConcurrentCommands),
		requestTimeout:          requestTimeout,
		semaphoreAcquireTimeout: semaphoreAcquireTimeout,
		idToCommandInfo:         idToCommandInfo,
	}
}

func (runCommandsHandler *runCommandsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id := params.ByName("id")
	commandInfo, ok := runCommandsHandler.idToCommandInfo[id]

	if !ok {
		slog.Warn("RunCommandsHandler unable to find comand",
			"id", id)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	runCommandsHandler.handleRunCommandRequest(&commandInfo, w, r)
}

func (runCommandsHandler *runCommandsHandler) handleRunCommandRequest(
	commandInfo *commandInfoDTO,
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, cancel := context.WithTimeout(r.Context(), runCommandsHandler.requestTimeout)
	defer cancel()

	commandAPIResponse := runCommandsHandler.runCommand(ctx, commandInfo)

	jsonText, err := json.Marshal(commandAPIResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add(utils.ContentTypeHeaderKey, utils.ContentTypeApplicationJSON)
	io.Copy(w, bytes.NewReader(jsonText))
}

func (runCommandsHandler *runCommandsHandler) acquireCommandSemaphore(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, runCommandsHandler.semaphoreAcquireTimeout)
	defer cancel()

	err = runCommandsHandler.commandSemaphore.Acquire(ctx, 1)
	if err != nil {
		err = fmt.Errorf("commandHandler.acquireCommandSemaphore error calling Acquire: %w", err)
	}
	return
}

func (runCommandsHandler *runCommandsHandler) releaseCommandSemaphore() {
	runCommandsHandler.commandSemaphore.Release(1)
}

type commandAPIResponse struct {
	CommandInfo                 *commandInfoDTO `json:"command_info"`
	Now                         string          `json:"now"`
	CommandDurationMilliseconds int64           `json:"command_duration_ms"`
	CommandOutput               string          `json:"command_output"`
}

func (runCommandsHandler *runCommandsHandler) runCommand(ctx context.Context, commandInfo *commandInfoDTO) commandAPIResponse {
	err := runCommandsHandler.acquireCommandSemaphore(ctx)
	if err != nil {
		return commandAPIResponse{
			CommandInfo:   commandInfo,
			Now:           utils.FormatTime(time.Now()),
			CommandOutput: fmt.Sprintf("%v", err),
		}
	}
	defer runCommandsHandler.releaseCommandSemaphore()

	commandStartTime := time.Now()
	rawCommandOutput, err := exec.CommandContext(
		ctx, commandInfo.Command, commandInfo.Args...).CombinedOutput()
	commandEndTime := time.Now()

	commandDuration := commandEndTime.Sub(commandStartTime)

	var commandOutput string
	if err != nil {
		commandOutput = fmt.Sprintf("command error %v", err)
	} else {
		commandOutput = string(rawCommandOutput)
	}

	return commandAPIResponse{
		CommandInfo:                 commandInfo,
		Now:                         utils.FormatTime(commandEndTime),
		CommandDurationMilliseconds: commandDuration.Milliseconds(),
		CommandOutput:               commandOutput,
	}
}
