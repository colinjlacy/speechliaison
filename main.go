package main

import (
	"fmt"
	"github.com/arienmalec/alexa-go"
	"github.com/aws/aws-lambda-go/lambda"
	"os"
	"speechLiason/cloud_resources"
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
	var r alexa.Response
	switch request.Body.Intent.Name {
	case "sync":
		c := request.Body.Intent.Slots["spokenCode"].Value
		t := request.Context.System.APIAccessToken
		d := request.Context.System.Device.DeviceID
		r = Sync(c, u, t, d)
		break
	case "createJob":
		j := request.Body.Intent.Slots["jobName"].Value
		r = CreateJob(j, u)
		break
	case "scan":
		j := request.Body.Intent.Slots["jobName"].Value
		_,_ =fmt.Fprintf(os.Stdout, "going to scan with user %s, and jobname %s", u, j)
		r = Scan(j, u)
		break
	case "emailJob":
		j := request.Body.Intent.Slots["jobName"].Value
		r = Deliver(j, u)
		break
	case "AMAZON.HelpIntent":
		r = respond.Welcome()
		break
	case "AMAZON.FallbackIntent":
		r = respond.Welcome()
		break
	default:
		r = respond.Welcome()
		break
	}
	return r, nil
}

func Sync(code string, voiceUserId string, token string, deviceId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err)
	}
	defer queue_connect.CloseConnection()
	addr, err := cloud_resources.GetDeviceAddress(token, deviceId)
	if err != nil {
		return errors.AnalyzeError(err)
	}
	err = queue_connect.SyncAccounts(code, voiceUserId, addr)
	if err != nil {
		return errors.AnalyzeError(err)
	}
	return respond.Positively("sync your account", false)
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
