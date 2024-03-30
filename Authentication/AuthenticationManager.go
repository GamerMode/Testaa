package Authentication

import (
	"client/ClientErrors"
	"client/Requests"
	"encoding/json"
	"fmt"
	"net"
)

const (
	signupArguments = 3
	loginArguments  = 2

	username_index = 0
	password_index = 1
	email_index    = 2
)

// function, argumentCount, arguments,

// Handles the sign up request
func HandleSignup(commandArguments []string, socket *net.Conn) error {
	if len(commandArguments) != signupArguments { // if Signup fields was not provided
		return &(ClientErrors.InvalidArgumentCountError{Arguments: uint8(len(commandArguments)), Expected: uint8(signupArguments)})
	}
	user := Signup(commandArguments[username_index], commandArguments[password_index], commandArguments[email_index]) // Signup a user struct
	request_data, err := json.Marshal(user)                                                                           // Convert user struct
	if err != nil {
		return &ClientErrors.JsonEncodeError{}
	}
	_, err = Requests.SendRequest(Requests.SignupRequest, request_data, socket) // Sends sign up request

	return err

}

// Handles the sign in request
func HandleSignIn(command_arguments []string, socket *net.Conn) error {
	if len(command_arguments) != 2 { // If username and password was not provided
		return fmt.Errorf("incorrect number of arguments.\nPlease try again")
	}
	user := Signin(command_arguments[username_index], command_arguments[password_index]) //Sign in a user struct
	request_data, err := json.Marshal(user)                                              // Convert user struct to raw json bytes
	if err != nil {
		return &ClientErrors.JsonEncodeError{}
	}
	_, err = Requests.SendRequest(Requests.LoginRequest, request_data, socket) // Sends sign in request

	return err
}
