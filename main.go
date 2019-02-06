package main

import (
	"fmt"
	"github.com/arienmalec/alexa-go"
	"github.com/aws/aws-lambda-go/lambda"
	"os"
	"speechLiason/errors"
	"speechLiason/key_access"
	"speechLiason/queue_connect"
	"speechLiason/respond"
)

func main() {
	lambda.Start(DispatchIntents)
}

func DispatchIntents(request alexa.Request) (alexa.Response, error) {
	u := request.Session.User.UserID
	j := request.Body.Intent.Slots["jobName"].Value
	var response alexa.Response
	switch request.Body.Intent.Name {
	case "createJob":
		response = CreateJob(j, u)
		break
	case "scan":
		_,_ =fmt.Fprintf(os.Stdout, "going to scan with user %s, and jobname %s", u, j)
		response = Scan(j, u)
		break
	case "emailJob":
		response = Deliver(j, u)
		break
	case "AMAZON.HelpIntent":
		response = respond.Welcome()
		break
	case "AMAZON.FallbackIntent":
		response = respond.Welcome()
		break
	default:
		response = respond.Welcome()
		break
	}
	return response, nil
}

func Scan(jobName, voiceUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err)
	}
	defer queue_connect.CloseConnection()
	if err := queue_connect.SendScanCommand(jobName, voiceUserId); err != nil {
		return errors.AnalyzeError(err)
	}
	return respond.Positively("scan a page", false)
}

func CreateJob(jobName, voiceUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err)
	}
	defer queue_connect.CloseConnection()
	if err := queue_connect.SetCursor(jobName, voiceUserId); err != nil {
		return errors.AnalyzeError(err)
	}
	return respond.Positively("create a job", false)
}

func Deliver(jobName, voiceUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err)
	}
	defer queue_connect.CloseConnection()
	if err := queue_connect.SendDeliveryCommand(jobName, voiceUserId, "email", "colinjlacy@gmail.com"); err != nil {
		return errors.AnalyzeError(err)
	}
	return respond.Positively("email job " + jobName, false)
}
