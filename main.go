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

/*
TODO:
- handle default errors better
- figure out what happens when user says "no" in response to code confirmation
- figure out better way to give code without multiple prompts
*/

func DispatchIntents(request alexa.Request) (alexa.Response, error) {
	var s string
	var p string
	u := request.Session.User.UserID
	if request.Session.Attributes["syncUserId"] != nil {
		s = request.Session.Attributes["syncUserId"].(string)
	}
	if request.Session.Attributes["previousJobCursor"] != nil {
		p = request.Session.Attributes["previousJobCursor"].(string)
	}
	var r alexa.Response
	fmt.Print("request body", request.Body)
	fmt.Print("request session", request.Session)
	switch request.Body.Intent.Name {
	case "sync":
		c := request.Body.Intent.Slots["spokenCode"].Value
		t := request.Context.System.APIAccessToken
		d := request.Context.System.Device.DeviceID
		r = Sync(c, u, t, d)
		break
	case "createJob":
		j := request.Body.Intent.Slots["jobName"].Value
		if j == "" {
			j = p
		}
		r = CreateJob(j, u, s)
		break
	case "scan":
		j := request.Body.Intent.Slots["jobName"].Value
		if j == "" {
			j = p
		}
		_, _ = fmt.Fprintf(os.Stdout, "going to scan with user %s, and jobname %s", u, j)
		r = Scan(j, u, s)
		break
	case "emailJob":
		t := request.Context.System.APIAccessToken
		j := request.Body.Intent.Slots["jobName"].Value
		if j == "" {
			j = p
		}
		r = Deliver(t, j, u, s)
		break
	case "scanAndEmail":
		t := request.Context.System.APIAccessToken
		r = QuickScanAndSend(t, u, s)
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
		return errors.AnalyzeError(err, "", "")
	}
	defer queue_connect.CloseConnection()
	addr, err := cloud_resources.GetDeviceAddress(token, deviceId)
	if err != nil {
		return errors.AnalyzeError(err, "", "")
	}
	err = queue_connect.SyncAccounts(code, voiceUserId, addr)
	if err != nil {
		return errors.AnalyzeError(err, "", "")
	}
	return respond.Positively("sync your account", false, "", "")
}

func Scan(jobName, voiceUserId, possibleSyncUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err, possibleSyncUserId, jobName)
	}
	defer queue_connect.CloseConnection()
	u, j, err := queue_connect.SendScanCommand(jobName, voiceUserId, possibleSyncUserId)
	if err != nil {
		return errors.AnalyzeError(err, u, j)
	}
	return respond.Positively("scan a page", false, u, j)
}

func CreateJob(jobName, voiceUserId, possibleSyncUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err, possibleSyncUserId, jobName)
	}
	defer queue_connect.CloseConnection()
	u, j, err := queue_connect.SetCursor(jobName, voiceUserId, possibleSyncUserId)
	if err != nil {
		return errors.AnalyzeError(err, u, j)
	}
	return respond.Positively("create a job", false, u, j)
}

func Deliver(token, jobName, voiceUserId, possibleSyncUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err, possibleSyncUserId, jobName)
	}
	defer queue_connect.CloseConnection()
	email, err := cloud_resources.GetUserEmail(token, voiceUserId)
	if err != nil {
		return errors.AnalyzeError(err, possibleSyncUserId, jobName)
	}
	u, j, err := queue_connect.SendDeliveryCommand(jobName, voiceUserId, possibleSyncUserId, "email", email)
	if err != nil {
		return errors.AnalyzeError(err, u, j)
	}
	return respond.Positively("email the job "+jobName, false, u, j)
}

func QuickScanAndSend(token, voiceUserId, possibleSyncUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err, "", "")
	}
	defer queue_connect.CloseConnection()
	email, err := cloud_resources.GetUserEmail(token, voiceUserId)
	if err != nil {
		return errors.AnalyzeError(err, "", "")
	}
	u, err := queue_connect.QuickScanAndDeliver(voiceUserId, possibleSyncUserId, "email", email)
	if err != nil {
		return errors.AnalyzeError(err, u, "")
	}
	return respond.Positively("scan and email this file to you", false, u, "")
}
