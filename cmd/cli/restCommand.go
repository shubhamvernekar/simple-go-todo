package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/urfave/cli/v2"
)

var (
	ErrInvalidID     = errors.New("invalid id")
	ErrInvalidTitle  = errors.New("invalid title")
	ErrJSONUnmarshal = errors.New("json unmarshal")
	ErrServerError   = errors.New("server error")
)

var APIFlags = []cli.Flag{
	&cli.StringFlag{
		Name: "title",
	},
	&cli.IntFlag{
		Name: "id",
	},
}

var APICommands = []*cli.Command{
	{
		Name:   "get_all",
		Usage:  "Get all todos list",
		Action: getAllTodos,
	},
	{
		Name:   "get",
		Usage:  "Get specific todos by id",
		Action: getTodo,
	},
	{
		Name:   "create",
		Usage:  "Create todo",
		Action: createTodo,
	},
	{
		Name:   "mark_done",
		Usage:  "Toggle done status of todo by id",
		Action: markDone,
	},
	{
		Name:   "delete",
		Usage:  "delete todo by given id",
		Action: deleteTodo,
	},
}

type RequestCaller interface {
	makeHTTPCall(requestType string, route string, jsonString string) ([]byte, error)
}

type HTTPRequestCaller struct {
	URL string
}

func (h *HTTPRequestCaller) makeHTTPCall(requestType string, route string, jsonString string) ([]byte, error) {
	client := &http.Client{}
	url := (h.URL + route)

	var bodyReader io.Reader

	if len(jsonString) > 0 {
		jsonBody := []byte(jsonString)
		bodyReader = bytes.NewReader(jsonBody)
	}

	request, err := http.NewRequest(requestType, url, bodyReader)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w http statusCode %d, responseData %s", ErrServerError, response.StatusCode, responseData)
	}

	return responseData, nil
}

var requestCaller RequestCaller

type Todo struct {
	ID     int    `json:"id"`
	Desc   string `json:"desc"`
	IsDone bool   `json:"is_done"`
}

type Todos []struct {
	Todo
}

func getAllTodos(_ *cli.Context) error {
	if requestCaller == nil {
		requestCaller = &HTTPRequestCaller{
			URL: cfg.BaseURL + cfg.Port,
		}
	}

	responseData, err := requestCaller.makeHTTPCall(http.MethodGet, GetAllTodo, "")
	if err != nil {
		return fmt.Errorf("%w : %w", ErrServerError, err)
	}

	todos := Todos{}
	err = json.Unmarshal(responseData, &todos)

	if err != nil {
		return fmt.Errorf("%w : %w", ErrJSONUnmarshal, err)
	}

	PrintTable(todos)
	return nil
}

func getTodo(ctx *cli.Context) error {
	id := ctx.Int("id")

	if id == 0 {
		return fmt.Errorf("failed to parse id %w", ErrInvalidID)
	}

	url := fmt.Sprintf("%s/%d", GetTodo, id)

	if requestCaller == nil {
		requestCaller = &HTTPRequestCaller{
			URL: cfg.BaseURL + cfg.Port,
		}
	}

	responseData, err := requestCaller.makeHTTPCall(http.MethodGet, url, "")
	if err != nil {
		return fmt.Errorf("%w : %w", ErrServerError, err)
	}

	var todo Todo
	err = json.Unmarshal(responseData, &todo)

	if err != nil {
		return fmt.Errorf("%w : %w", ErrJSONUnmarshal, err)
	}

	todos := Todos{struct{ Todo }{todo}}
	PrintTable(todos)
	return nil
}

func createTodo(ctx *cli.Context) error {
	title := ctx.String("title")

	if len(title) == 0 {
		return fmt.Errorf("failed to parse title %w ", ErrInvalidTitle)
	}

	jsonString := fmt.Sprintf(`{"desc":"%s"}`, title)

	if requestCaller == nil {
		requestCaller = &HTTPRequestCaller{
			URL: cfg.BaseURL + cfg.Port,
		}
	}

	responseData, err := requestCaller.makeHTTPCall(http.MethodPost, PostTodo, jsonString)
	if err != nil {
		return fmt.Errorf("%w : %w", ErrServerError, err)
	}

	todo := Todo{}
	err = json.Unmarshal(responseData, &todo)

	if err != nil {
		return fmt.Errorf("%w : %w", ErrJSONUnmarshal, err)
	}

	todos := Todos{struct{ Todo }{todo}}
	PrintTable(todos)
	return nil
}

func deleteTodo(ctx *cli.Context) error {
	id := ctx.Int("id")

	if id == 0 {
		return fmt.Errorf("falied to parse id %w ", ErrInvalidID)
	}

	url := fmt.Sprintf("%s/%d", DeleteTodo, id)

	if requestCaller == nil {
		requestCaller = &HTTPRequestCaller{
			URL: cfg.BaseURL + cfg.Port,
		}
	}

	_, err := requestCaller.makeHTTPCall(http.MethodDelete, url, "")
	if err != nil {
		return fmt.Errorf("%w : %w", ErrServerError, err)
	}

	log.Printf("Todo %d Deleted successfully", id)
	return nil
}

func markDone(ctx *cli.Context) error {
	id := ctx.Int("id")

	if id == 0 {
		return fmt.Errorf("falied to parse id %w ", ErrInvalidID)
	}

	url := fmt.Sprintf("%s/%d", MarkDone, id)

	if requestCaller == nil {
		requestCaller = &HTTPRequestCaller{
			URL: cfg.BaseURL + cfg.Port,
		}
	}

	responseData, err := requestCaller.makeHTTPCall(http.MethodGet, url, "")
	if err != nil {
		return fmt.Errorf("%w : %w", ErrServerError, err)
	}

	todo := Todo{}
	err = json.Unmarshal(responseData, &todo)
	if err != nil {
		return fmt.Errorf("%w : %w", ErrJSONUnmarshal, err)
	}

	todos := Todos{struct{ Todo }{todo}}
	PrintTable(todos)
	return nil
}
