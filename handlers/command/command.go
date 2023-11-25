package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"slices"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/sync/semaphore"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/utils"
)

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
		Args:        slices.Clone(commandInfo.Args),
	}
}

func NewAllCommandsHandler(commandConfiguration config.CommandConfiguration) (http.Handler, error) {
	allCommandDTOs := make([]commandInfoDTO, 0, len(commandConfiguration.Commands))
	for _, command := range commandConfiguration.Commands {
		allCommandDTOs = append(allCommandDTOs, commandInfoToDTO(command))
	}

	jsonBytes, err := json.Marshal(allCommandDTOs)
	if err != nil {
		return nil, fmt.Errorf("NewAllCommandsHandler json.Marshal error: %w", err)
	}

	return utils.JSONBytesHandlerFunc(jsonBytes), nil
}

type runCommandsHandler struct {
	commandSemaphore        *semaphore.Weighted
	requestTimeout          time.Duration
	semaphoreAcquireTimeout time.Duration
	idToCommandInfo         map[string]commandInfoDTO
}

func NewRunCommandsHandler(commandConfiguration config.CommandConfiguration) (http.Handler, error) {
	requestTimeout, err := time.ParseDuration(commandConfiguration.RequestTimeoutDuration)
	if err != nil {
		return nil, fmt.Errorf("NewRunCommandsHandler error parsing RequestTimeoutDuration: %w", err)
	}

	semaphoreAcquireTimeout, err := time.ParseDuration(commandConfiguration.SemaphoreAcquireTimeoutDuration)
	if err != nil {
		return nil, fmt.Errorf("NewRunCommandsHandler error parsing SemaphoreAcquireTimeoutDuration: %w", err)
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
	}, nil
}

func (runCommandsHandler *runCommandsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(ctx)

	id := params.ByName("id")
	commandInfo, ok := runCommandsHandler.idToCommandInfo[id]

	if !ok {
		slog.Warn("RunCommandsHandler unable to find comand",
			"id", id,
		)
		utils.HTTPErrorStatusCode(w, http.StatusNotFound)
		return
	}

	runCommandsHandler.handleRunCommandRequest(ctx, commandInfo, w)
}

func (runCommandsHandler *runCommandsHandler) handleRunCommandRequest(
	ctx context.Context,
	commandInfo commandInfoDTO,
	w http.ResponseWriter,
) {
	ctx, cancel := context.WithTimeout(ctx, runCommandsHandler.requestTimeout)
	defer cancel()

	commandAPIResponse, err := runCommandsHandler.runCommand(ctx, commandInfo)

	if err != nil {
		slog.Warn("RunCommandsHandler.runCommand returned error",
			"error", err,
		)
		switch {
		case errors.Is(err, errorAcquiringCommandSemaphore):
			utils.HTTPErrorStatusCode(w, http.StatusTooManyRequests)

		default:
			utils.HTTPErrorStatusCode(w, http.StatusInternalServerError)
		}
		return
	}

	utils.RespondWithJSONDTO(commandAPIResponse, w)
}

var errorAcquiringCommandSemaphore = errors.New("error acquiring command semaphore")

func (runCommandsHandler *runCommandsHandler) acquireCommandSemaphore(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, runCommandsHandler.semaphoreAcquireTimeout)
	defer cancel()

	err := runCommandsHandler.commandSemaphore.Acquire(ctx, 1)
	if err != nil {
		return errors.Join(errorAcquiringCommandSemaphore, err)
	}
	return nil
}

func (runCommandsHandler *runCommandsHandler) releaseCommandSemaphore() {
	runCommandsHandler.commandSemaphore.Release(1)
}

type commandAPIResponse struct {
	CommandInfo                 commandInfoDTO `json:"command_info"`
	Now                         string         `json:"now"`
	CommandDurationMilliseconds int64          `json:"command_duration_ms"`
	CommandOutput               string         `json:"command_output"`
}

func (runCommandsHandler *runCommandsHandler) runCommand(
	ctx context.Context,
	commandInfo commandInfoDTO,
) (response commandAPIResponse, err error) {
	err = runCommandsHandler.acquireCommandSemaphore(ctx)
	if err != nil {
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

	response = commandAPIResponse{
		CommandInfo:                 commandInfo,
		Now:                         utils.FormatTime(commandEndTime),
		CommandDurationMilliseconds: commandDuration.Milliseconds(),
		CommandOutput:               commandOutput,
	}
	err = nil
	return
}
