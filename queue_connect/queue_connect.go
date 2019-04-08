package queue_connect

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"os"
	"regexp"
	"speechLiason/cloud_resources"
	"speechLiason/errors"
	"time"
)

var client *firestore.Client
var ctx context.Context

const cursorTtl = 5 * time.Minute
const syncTtl = 3 * time.Minute

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

type syncDoc struct {
	VoiceUserId       string `firestore:"v"`
	VoiceUserLocation string `firestore:"l"`
	SpokenCode        string `firestore:"c"`
	GeneratedCode     string `firestore:"g"`
	Accepted          bool   `firestore:"a"`
	Initialized       int64  `firestore:"t"`
	UserId            string `firestore:"u"`
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

func SendScanCommand(jobName string, voiceUserId, possibleSyncUserId string) (syncUserId, jobCursor string, err error) {
	j, err := setJobName(jobName, voiceUserId)
	if err != nil {
		return "", "", err
	}
	u := possibleSyncUserId
	if u == "" {
		// TODO: set user ID
	}
	s := scanDoc{"", voiceUserId, j}
	if _, _, err := client.Collection("scan").Add(ctx, s); err != nil {
		return u, j, errors.SystemError{JobName: j, UserId: voiceUserId, Context: "SendScanCommand", Log: fmt.Sprintf("there was a problem creating the scan command: %s", err)}
	}
	_, _, err = SetCursor(j, voiceUserId, u)
	if err != nil {
		return u, j, err
	}
	return u, j, nil
}

func SendDeliveryCommand(jobName, voiceUserId, possibleSyncUserId, method, destination string) (syncUserId, jobCursor string, err error) {
	j, err := setJobName(jobName, voiceUserId)
	if err != nil {
		return possibleSyncUserId, "", err
	}
	if method != "email" {
		return possibleSyncUserId, j, errors.UnsupportedOperationError{ContextualError: errors.ContextualError{JobName: j, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("method for delivery %s not yet supported", method)}, ErroneousOperation: method}
	}
	if !isValidEmail(destination) {
		return possibleSyncUserId, j, errors.InvalidInputError{ContextualError: errors.ContextualError{JobName: j, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("invalid email address %s", destination)}, ErroneousInput: "email address"}
	}
	u := possibleSyncUserId
	if u == "" {
		// TODO: set user ID
	}
	d := deliverDoc{u, voiceUserId, j, method, destination}
	_, err = client.Doc("delivery").Create(ctx, d)
	if err != nil {
		return u, j, errors.SystemError{JobName: j, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("could not create delivery command: %s", err)}
	}
	return u, j, nil
}

func SetCursor(jobName, voiceUserId, possibleSyncUserId string) (syncUserId, jobCursor string, err error) {
	if jobName == "" {
		return possibleSyncUserId, "", errors.MissingJobNameError{JobName: jobName, UserId: voiceUserId, Context: "SetCursor", Log: "could not set cursor without a job name"}
	}
	ref := client.Collection("cursor").Doc(voiceUserId)
	u := possibleSyncUserId
	if u == "" {
		// TODO: set user ID
	}
	c := cursorDoc{UserId: u, Set: time.Now(), JobName: jobName}
	_, err = ref.Set(ctx, c)
	if err != nil {
		err = errors.SystemError{JobName: jobName, UserId: voiceUserId, Context: "SetCursor", Log: fmt.Sprintf("could not set new cursor doc: %s", err)}
	}
	return u, jobName, err
}

func SyncAccounts(spokenCode string, voiceUserId string, deviceAddress cloud_resources.DeviceAddress) (err error) {
	s, err := getSyncDoc(spokenCode)
	if err != nil {
		return err
	}
	if err = checkSyncDocExpired(s); err != nil {
		return err
	}
	s.VoiceUserId = voiceUserId
	s.SpokenCode = spokenCode
	s.VoiceUserLocation = deviceAddress.PromptedLocation
	err = setSyncDoc(s, s.UserId)
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

func setUserId(voiceUserId string) (userId string, err error) {
	// check session attributes
	// check dynamodb for user mapping
	// check for completed sync doc
	return
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
	return
}

func isCursorExpired(c cursorDoc) bool {
	s := time.Since(c.Set)
	return s.Minutes() > cursorTtl.Minutes()
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return re.MatchString(email)
}

func getSyncDoc(spokenCode string) (syncDoc, error) {
	q := client.Collection("sync").
		Where("g", "==", spokenCode).
		Where("a", "==", false)
	iter := q.Documents(ctx)
	s := syncDoc{}
	defer iter.Stop()
	/*
		purposely choosing the last doc in the iterator, as it's
		just as likely to be the correct one as the first doc;
		ideally there would only be one in the iterator.
		TODO: figure out a better way...
	*/
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return syncDoc{}, errors.SystemError{JobName: "", UserId: "", Context: "getSyncDoc", Log: fmt.Sprintf("unexpected error while retrieving sync doc: %s", err)}
		}
		err = doc.DataTo(&s)
		if err != nil {
			return syncDoc{}, errors.SystemError{JobName: "", UserId: "", Context: "getSyncDoc", Log: fmt.Sprintf("could not map sync doc to struct: %s", err)}
		}
	}
	if s == (syncDoc{}) {
		return syncDoc{}, errors.SyncDocNotFoundError{SpokenCode: spokenCode, Context: "getSyncDoc", Log: "cursor doc snapshot doesn't exist"}
	}
	_, _ = fmt.Fprintf(os.Stdout, "syncDoc: %v", s)
	return s, nil
}

func setSyncDoc(doc syncDoc, syncUserId string) error {
	ref := client.Collection("sync").Doc(syncUserId)
	_, err := ref.Set(ctx, doc)
	if err != nil {
		return errors.SystemError{SpokenCode: doc.SpokenCode, UserId: doc.VoiceUserId, Context: "setSyncDoc", Log: "error updating sync doc"}
	}
	return nil
}

func checkSyncDocExpired(s syncDoc) error {
	t := time.Now().UTC().Add(syncTtl * -1)
	i := time.Unix(s.Initialized / 1000, 0)
	_,_ = fmt.Fprintf(os.Stdout, "current time: %s; initialized time: %s", t, i)
	if i.After(t) {
		return nil
	}
	return errors.SyncDocExpiredError{SpokenCode: s.SpokenCode, Context: "checkSyncDocExpired", Log: fmt.Sprintf("sync doc expired; time initialized %s, expiration time %s", s.Initialized, t)}
}
