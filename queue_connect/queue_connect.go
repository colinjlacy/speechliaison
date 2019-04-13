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
	Method      string `firestore:"m"`
	Destination string `firestore:"d"`
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

func SendScanCommand(jobName string, voiceUserId, possibleSyncUserId string) (syncUserId, sessionJobName string, err error) {
	sessionJobName, err = setJobName(jobName, voiceUserId)
	if err != nil {
		return
	}
	syncUserId, err = getUserId(voiceUserId, possibleSyncUserId)
	if err != nil {
		return
	}
	s := scanDoc{syncUserId, voiceUserId, sessionJobName,"", ""}
	if _, _, err := client.Collection("scan").Add(ctx, s); err != nil {
		return syncUserId, sessionJobName, errors.SystemError{JobName: sessionJobName, UserId: syncUserId, Context: "SendScanCommand", Log: fmt.Sprintf("there was a problem creating the scan command: %s", err)}
	}
	_, _, err = SetCursor(sessionJobName, voiceUserId, syncUserId)
	return
}

// TODO: fix delivery, log output [could not create delivery command: firestore: nil DocumentRef]
func SendDeliveryCommand(jobName, voiceUserId, possibleSyncUserId, method, destination string) (syncUserId, sessionJobName string, err error) {
	sessionJobName, err = setJobName(jobName, voiceUserId)
	if err != nil {
		return possibleSyncUserId, "", err
	}
	if method != "email" {
		return possibleSyncUserId, sessionJobName, errors.UnsupportedOperationError{ContextualError: errors.ContextualError{JobName: sessionJobName, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("method for delivery %s not yet supported", method)}, ErroneousOperation: method}
	}
	if !isValidEmail(destination) {
		return possibleSyncUserId, sessionJobName, errors.InvalidInputError{ContextualError: errors.ContextualError{JobName: sessionJobName, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("invalid email address %s", destination)}, ErroneousInput: "email address"}
	}
	syncUserId, err = getUserId(voiceUserId, possibleSyncUserId)
	if err != nil {
		return
	}
	d := deliverDoc{syncUserId, voiceUserId, sessionJobName, method, destination}
	_, _, err = client.Collection("delivery").Add(ctx, d)
	if err != nil {
		return syncUserId, sessionJobName, errors.SystemError{JobName: sessionJobName, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("could not create delivery command: %s", err)}
	}
	return
}

func SetCursor(jobName, voiceUserId, possibleSyncUserId string) (syncUserId, sessionJobName string, err error) {
	if jobName == "" {
		return possibleSyncUserId, "", errors.MissingJobNameError{JobName: jobName, UserId: voiceUserId, Context: "SetCursor", Log: "could not set cursor without a job name"}
	}
	sessionJobName = jobName
	ref := client.Collection("cursor").Doc(voiceUserId)
	syncUserId, err = getUserId(voiceUserId, possibleSyncUserId)
	if err != nil {
		return
	}
	c := cursorDoc{UserId: syncUserId, Set: time.Now(), JobName: jobName}
	_, err = ref.Set(ctx, c)
	if err != nil {
		err = errors.SystemError{JobName: jobName, UserId: voiceUserId, Context: "SetCursor", Log: fmt.Sprintf("could not set new cursor doc: %s", err)}
	}
	return
}

func SyncAccounts(spokenCode string, voiceUserId string, deviceAddress cloud_resources.DeviceAddress) (err error) {
	s, err := getSyncDoc(spokenCode, "")
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

func QuickScanAndDeliver(voiceUserId, possibleSyncUserId, method, destination string) (syncUserId string, err error) {
	t := string(time.Now().Unix())
	syncUserId, err = getUserId(voiceUserId, possibleSyncUserId)
	if err != nil {
		return
	}
	if method != "email" {
		return "", errors.UnsupportedOperationError{ContextualError: errors.ContextualError{JobName: t, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("method for delivery %s not yet supported", method)}, ErroneousOperation: method}
	}
	if !isValidEmail(destination) {
		return "", errors.InvalidInputError{ContextualError: errors.ContextualError{JobName: t, UserId: voiceUserId, Context: "SendDeliveryCommand", Log: fmt.Sprintf("invalid email address %s", destination)}, ErroneousInput: "email address"}
	}
	s := scanDoc{syncUserId, voiceUserId, t,method, destination}
	if _, _, err := client.Collection("scan").Add(ctx, s); err != nil {
		return syncUserId, errors.SystemError{JobName: t, UserId: syncUserId, Context: "SendScanCommand", Log: fmt.Sprintf("there was a problem creating the scan command: %s", err)}
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

func getUserId(voiceUserId, possibleSyncUserId string) (userId string, err error) {
	// check session attributes
	if possibleSyncUserId != "" {
		userId = possibleSyncUserId
		return
	}
	// check dynamodb for user mapping
	um, err := cloud_resources.GetUserMapping(voiceUserId)
	if err != nil {
		return
	}
	if um != &(cloud_resources.UserMapping{}) {
		userId = um.SyncUserId
		return
	}
	// check for completed sync doc
	s, err := getSyncDoc("", voiceUserId)
	if err != nil {
		switch err.(type) {
		case errors.SyncDocNotFoundError:
			err = errors.UserAccountNotSyncedError{Context: "getUserId", Log: fmt.Sprintf("voice user %s has not yet synced device")}
			break
		default:
			break
		}
		return
	}
	userId = s.UserId
	if err = cloud_resources.SaveUserMapping(voiceUserId, userId); err != nil {
		e := errors.SystemError{UserId:userId, Context:"getUserId", Log:fmt.Sprintf("could not persist sync doc's info to user-mapping database: %s", err)}
		fmt.Println(e)
		return userId, nil
	}
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
	re := regexp.MustCompile(`^"[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}"$`)
	return re.MatchString(email)
}

func getSyncDoc(spokenCode, voiceUserId string) (syncDoc, error) {
	q, _ := getQuery(spokenCode, voiceUserId)
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
		return syncDoc{}, errors.SyncDocNotFoundError{SpokenCode: spokenCode, Context: "getSyncDoc", Log: fmt.Sprintf("cursor doc snapshot doesn't exist for code %s or userId %s", spokenCode, voiceUserId)}
	}
	_, _ = fmt.Fprintf(os.Stdout, "syncDoc: %v", s)
	return s, nil
}

func getQuery(syncCode, voiceUserId string) (firestore.Query, error) {
	if syncCode != "" {
		return client.Collection("sync").
			Where("g", "==", syncCode).
			Where("a", "==", false), nil
	}
	if voiceUserId != "" {
		return client.Collection("sync").
			Where("v", "==", voiceUserId).
			Where("a", "==", true), nil
	}
	return firestore.Query{}, errors.SystemError{Context: "getQuery", Log: "you need to provide either a syncCode or a voiceUserId!"}
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
