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

func CreateAllCommandsHandler(commandConfiguration config.CommandConfiguration) httprouter.Handle {
	jsonBuffer, err := json.Marshal(commandConfiguration.Commands)
	if err != nil {
		slog.Error("getAllCommandsHandlerFunc json.Marshal error",
			"error", err)
		os.Exit(1)
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Add(utils.ContentTypeHeaderKey, utils.ContentTypeApplicationJSON)
		io.Copy(w, bytes.NewReader(jsonBuffer))
	}
}

type runCommandsHandler struct {
	commandSemaphore        *semaphore.Weighted
	requestTimeout          time.Duration
	semaphoreAcquireTimeout time.Duration
	idToCommandInfo         map[string]config.CommandInfo
}

func newRunCommandsHandler(commandConfiguration config.CommandConfiguration) *runCommandsHandler {
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

	idToCommandInfo := make(map[string]config.CommandInfo)
	for _, commandInfo := range commandConfiguration.Commands {
		idToCommandInfo[commandInfo.ID] = commandInfo
	}

	handler := &runCommandsHandler{
		commandSemaphore:        semaphore.NewWeighted(commandConfiguration.MaxConcurrentCommands),
		requestTimeout:          requestTimeout,
		semaphoreAcquireTimeout: semaphoreAcquireTimeout,
		idToCommandInfo:         idToCommandInfo,
	}

	return handler
}

func CreateRunCommandsHandler(commandConfiguration config.CommandConfiguration) httprouter.Handle {
	runCommandsHandler := newRunCommandsHandler(commandConfiguration)

	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		commandInfo, ok := runCommandsHandler.idToCommandInfo[id]

		if !ok {
			slog.Warn("RunCommandsHandler unable to find comand id",
				"id", id)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		runCommandsHandler.handleRunCommandRequest(commandInfo, w, r)
	}
}

func (runCommandsHandler *runCommandsHandler) handleRunCommandRequest(
	commandInfo config.CommandInfo,
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, cancel := context.WithTimeout(r.Context(), runCommandsHandler.requestTimeout)
	defer cancel()

	commandAPIResponse := runCommandsHandler.runCommand(ctx, &commandInfo)

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
	CommandInfo                 *config.CommandInfo `json:"command_info"`
	Now                         string              `json:"now"`
	CommandDurationMilliseconds int64               `json:"command_duration_ms"`
	CommandOutput               string              `json:"command_output"`
}

func (runCommandsHandler *runCommandsHandler) runCommand(ctx context.Context, commandInfo *config.CommandInfo) (response *commandAPIResponse) {
	err := runCommandsHandler.acquireCommandSemaphore(ctx)
	if err != nil {
		response = &commandAPIResponse{
			CommandInfo:   commandInfo,
			Now:           utils.FormatTime(time.Now()),
			CommandOutput: fmt.Sprintf("%v", err),
		}
		return
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

	response = &commandAPIResponse{
		CommandInfo:                 commandInfo,
		Now:                         utils.FormatTime(commandEndTime),
		CommandDurationMilliseconds: commandDuration.Milliseconds(),
		CommandOutput:               commandOutput,
	}
	return
}
