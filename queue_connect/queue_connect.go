package queue_connect

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"regexp"
	"speechLiason/errors"
	"time"
)

var client *firestore.Client
var ctx context.Context

const cursorTtl = 5 * time.Minute

type scanDoc struct {
	UserId      string `firestore:"u"`
	VoiceUserId string `firestore:"v"`
	JobName     string `firestore:"j"`
}

type cursorDoc struct {
	Set     time.Time `firestore:"s"`
	UserId  string    `firestore:"u"`
	JobName string    `firestore:"j"`
}

type deliverDoc struct {
	UserId      string `firestore:"u"`
	VoiceUserId string `firestore:"v"`
	JobName     string `firestore:"j"`
	Method      string `firestore:"m"`
	Destination string `firestore:"d"`
}

type userDoc struct {
	UserId string `firestore:"u"`
}

type authDoc struct {
	VoiceUserId       string    `firestore:"v"`
	VoiceUserLocation string    `firestore:"l"`
	AuthCode          int8      `firestore:"c"`
	Expires           time.Time `firestore:"e"`
}

func InitConnection(projectName string, key []byte) (err error) {
	ctx = context.Background()
	opts := option.WithCredentialsJSON(key)
	client, err = firestore.NewClient(ctx, projectName, opts)
	if err != nil {
		return errors.SystemError{JobName: projectName, UserId: "", Context: "InitConnection", Log: fmt.Sprintf("could not initialize connection: %s", err)}
	}
	return err
}

func CloseConnection() {
	_ = client.Close()
}

func SendScanCommand(jobName string, voiceUserId string) error {
	j, err := setJobName(jobName, voiceUserId)
	if err != nil {
		return err
	}
	u, err := getUserId(voiceUserId)
	if err != nil {
		return err
	}
	s := scanDoc{u, voiceUserId, j}
	if _, _, err := client.Collection("scan").Add(ctx, s); err != nil {
		return errors.SystemError{JobName:jobName, UserId:voiceUserId, Context:"SendScanCommand", Log:fmt.Sprintf("there was a problem creating the scan command: %s", err)}
	}
	if err := SetCursor(jobName, voiceUserId); err != nil {
		return err
	}
	return nil
}

func SendDeliveryCommand(jobName, voiceUserId, method, destination string) error {
	j, err := setJobName(jobName, voiceUserId)
	if err != nil {
		return err
	}
	if method != "email" {
		return errors.UnsupportedOperationError{ContextualError: errors.ContextualError{JobName: jobName, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("method for delivery %s not yet supported", method)}, ErroneousOperation: method}
	}
	if !isValidEmail(destination) {
		return errors.InvalidInputError{ContextualError: errors.ContextualError{JobName:jobName, UserId:voiceUserId, Context:"SendDeliveryCommand", Log:fmt.Sprintf("invalid email address %s", destination)}, ErroneousInput:"email address"}
	}
	d := deliverDoc{"", voiceUserId, j, method, destination}
	_, err = client.Doc("delivery").Create(ctx, d)
	if err != nil {
		return errors.SystemError{JobName:jobName, UserId:voiceUserId, Context:"SendDeliveryCommand", Log:fmt.Sprintf("could not create delivery command: %s", err)}
	}
	return nil
}

func SetCursor(jobName, voiceUserId string) (err error) {
	if jobName == "" {
		err = errors.MissingJobNameError{JobName: "", UserId: voiceUserId, Context: "SetCursor", Log: "could not set cursor without a job name"}
		return
	}
	ref := client.Collection("cursor").Doc(voiceUserId)
	u, err := getUserId(voiceUserId)
	if err != nil {
		return err
	}
	c := cursorDoc{UserId: u, Set: time.Now(), JobName: jobName}
	_, err = ref.Set(ctx, c)
	if err != nil {
		err = errors.SystemError{JobName: jobName, UserId: voiceUserId, Context: "SetCursor", Log: fmt.Sprintf("could not set new cursor doc: %s", err)}
	}
	return
}

func setJobName(inputName, voiceUserId string) (outputName string, err error) {
	if inputName != "" {
		return inputName, nil
	}
	c, err := getCursorDoc(voiceUserId)
	if err != nil {
		return "", err
	}
	if c == (cursorDoc{}) {
		return "", errors.CursorNotFoundError{JobName: "", UserId: voiceUserId, Context: "setJobName", Log: "no error retrieving cursor doc, but it came back as nil"}
	}
	outputName = c.JobName
	if isCursorExpired(c) {
		return "", errors.CursorExpiredError{JobName: outputName, UserId: voiceUserId, Context: "setJobName", Log: "cursor is expired"}
	}
	return
}

func getUserId(voiceUserId string) (userId string, err error) {
	q := client.Collection("usersMappings")
	ds, err := q.Documents(ctx).GetAll()
	if err != nil {
		return "", errors.SystemError{JobName: "", UserId: voiceUserId, Context: "getUserId", Log: fmt.Sprintf("could not get user document list: %s", err)}
	}
	var d userDoc
	for _, item := range ds {
		err = item.DataTo(&d)
		fmt.Printf("user doc: %+v", item.Data())
		break
	}
	if err != nil {
		return "", errors.SystemError{JobName: "", UserId: voiceUserId, Context: "getCursorDoc", Log: fmt.Sprintf("could not parse user document: %s", err)}
	}
	userId = d.UserId
	return
}

func isCursorExpired(c cursorDoc) bool {
	s := time.Since(c.Set)
	return s.Minutes() > cursorTtl.Minutes()
}

func getCursorDoc(voiceUserId string) (cursor cursorDoc, err error) {
	snap, err := client.Doc("cursor/" + voiceUserId).Get(ctx)
	if err != nil {
		return cursorDoc{}, errors.CursorNotFoundError{JobName: "", UserId: voiceUserId, Context: "getCursorDoc", Log: "cursor doc snapshot doesn't exist"}
	}
	if snap.Exists() != true {
		return cursorDoc{}, errors.CursorNotFoundError{JobName: "", UserId: voiceUserId, Context: "getCursorDoc", Log: "cursor doc snapshot doesn't exist"}
	}
	err = snap.DataTo(&cursor)
	if err != nil {
		return cursorDoc{}, errors.SystemError{JobName: "", UserId: voiceUserId, Context: "getCursorDoc", Log: "cursor doc snapshot doesn't exist"}
	}
	return cursor, nil
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return re.MatchString(email)
}
