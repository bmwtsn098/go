// Package tests has the unit tests of package messaging.
// common file has the reused methods across the varoius unit test files
package tests

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/anovikov1984/go-vcr/cassette"
	"github.com/stretchr/testify/assert"
)

// PamSubKey: key for pam tests
var PamSubKey = "sub-c-90c51098-c040-11e5-a316-0619f8945a4f"

// PamPubKey: key for pam tests
var PamPubKey = "pub-c-1bd448ed-05ba-4dbc-81a5-7d6ff5c6e2bb"

// PamSecKey: key for pam tests
var PamSecKey = "sec-c-ZDA1ZTdlNzAtYzU4Zi00MmEwLTljZmItM2ZhMDExZTE2ZmQ5"

// SubKey: key for non-pam tests
var SubKey = "sub-c-5c4fdcc6-c040-11e5-a316-0619f8945a4f"

// PubKey: key for non-pam tests
var PubKey = "pub-c-071e1a3f-607f-4351-bdd1-73a8eb21ba7c"

// SecKey: key for non-pam tests
var SecKey = "sec-c-ZjM0NzNmODgtNzE4OC00OTBjLWFhMWMtYjUxZTllYmY5YWE4"

// timeoutMessage is the text message displayed when the
// unit test times out
var timeoutMessage = "Test timed out."

// testTimeout in seconds
var testTimeout int = 30

// prefix for presence channels
var presenceSuffix string = "-pnpres"

// publishSuccessMessage: the reponse that is received when a message is
// successfully published on a pubnub channel.
var publishSuccessMessage = "1,\"Sent\""

// EmptyStruct provided the empty struct to test the encryption.
type EmptyStruct struct {
}

// CustomStruct to test the custom structure encryption and decryption
// The variables "foo" and "bar" as used in the other languages are not
// accepted by golang and give an empty value when serialized, used "Foo"
// and "Bar" instead.
type CustomStruct struct {
	Foo string
	Bar []int
}

// CustomSingleElementStruct Used to test the custom structure encryption and decryption
// The variables "foo" and "bar" as used in the other languages are not
// accepted by golang and give an empty value when serialized, used "Foo"
// and "Bar" instead.
type CustomSingleElementStruct struct {
	Foo string
}

// CustomComplexMessage is used to test the custom structure encryption and decryption.
// The variables "foo" and "bar" as used in the other languages are not
// accepted by golang and give an empty value when serialized, used "Foo"
// and "Bar" instead.
type CustomComplexMessage struct {
	VersionID     float32 `json:",string"`
	TimeToken     int64   `json:",string"`
	OperationName string
	Channels      []string
	DemoMessage   PubnubDemoMessage `json:",string"`
	SampleXML     string            `json:",string"`
}

// PubnubDemoMessage is a struct to test a non-alphanumeric message
type PubnubDemoMessage struct {
	DefaultMessage string `json:",string"`
}

// GenRandom gets a random instance
func GenRandom() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// InitComplexMessage initializes a complex structure of the
// type CustomComplexMessage which includes a xml, struct of type PubnubDemoMessage,
// strings, float and integer.
func InitComplexMessage() CustomComplexMessage {
	pubnubDemoMessage := PubnubDemoMessage{
		DefaultMessage: "~!@#$%^&*()_+ `1234567890-= qwertyuiop[]\\ {}| asdfghjkl;' :\" zxcvbnm,./ <>? ",
	}

	xmlDoc := &Data{Name: "Doe", Age: 42}

	//_, err := xml.MarshalIndent(xmlDoc, "  ", "    ")
	//output, err := xml.MarshalIndent(xmlDoc, "  ", "    ")
	output := new(bytes.Buffer)
	enc := xml.NewEncoder(output)

	err := enc.Encode(xmlDoc)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return CustomComplexMessage{}
	}
	//fmt.Printf("xmlDoc: %v\n", xmlDoc)
	customComplexMessage := CustomComplexMessage{
		VersionID:     3.4,
		TimeToken:     13601488652764619,
		OperationName: "Publish",
		Channels:      []string{"ch1", "ch 2"},
		DemoMessage:   pubnubDemoMessage,
		//SampleXml        : xmlDoc,
		SampleXML: output.String(),
	}
	return customComplexMessage
}

// Data represents a <data> element.
type Data struct {
	XMLName xml.Name `xml:"data"`
	//Entry   []Entry  `xml:"entry"`
	Name string `xml:"name"`
	Age  int    `xml:"age"`
}

// Entry represents an <entry> element.
type Entry struct {
	Name string `xml:"name"`
	Age  int    `xml:"age"`
}

type PamResponse struct {
	Payload interface{}
	Status  int
	Service string
	Message string
}

// PrintTestMessage is  common method to print the message on the screen.
func PrintTestMessage(message string) {
	fmt.Println(" ")
	fmt.Println(message)
	fmt.Println(" ")
}

// ReplaceEncodedChars takes a string as a parameter and returns a string
// with the unicode chars \\u003c, \\u003e, \\u0026  with <,> and & respectively
func ReplaceEncodedChars(str string) string {
	str = strings.Replace(str, "\\u003c", "<", -1)
	str = strings.Replace(str, "\\u003e", ">", -1)
	str = strings.Replace(str, "\\u0026", "&", -1)
	return str
}

// WaitForCompletion reads the response on the responseChannel or waits till the timeout
// occurs. if the response is received before the timeout the response is sent to the
// waitChannel else the test is timed out.
//
// Parameters:
// responseChannel: channel to read.
// waitChannel: channel to respond to.
func WaitForCompletion(responseChannel chan string, waitChannel chan string) {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(30 * time.Second)
		timeout <- true
	}()
	for {
		select {
		case value, ok := <-responseChannel:
			if !ok {
				break
			}

			if value != "[]" {
				waitChannel <- value
				timeout <- false
				//break
			}
			break
		case <-timeout:
			//case b, _ := <-timeout:
			//if b {
			waitChannel <- timeoutMessage
			//}
			break
		}
	}
}

// ParseWaitResponse parses the response of the wait channel.
// If the response contains the string "passed" then the test is passed else it is failed.
//
// Parameters:
// waitChannel: channel to read
// t: the testing.T instance
// testName to display.
func ParseWaitResponse(waitChannel chan string, t *testing.T, testName string) {
	for {
		value, ok := <-waitChannel
		if !ok {
			break
		}
		returnVal := string(value)
		if returnVal != "[]" {
			//fmt.Println("wait:", returnVal)
			if strings.Contains(returnVal, "passed") {
				//fmt.Println("Test '" + testName + "': passed.")
			} else {
				fmt.Println("Test '" + testName + "': failed. Message: " + returnVal)
				t.Error("Test '" + testName + "': failed.")
			}
			break
		}
	}
}

// ParseErrorResponse parses the response of the Error channel.
// It prints the response to the response channel
func ParseErrorResponse(channel chan []byte, responseChannel chan string) {
	for {
		value, ok := <-channel
		if !ok {
			break
		}
		returnVal := string(value)
		if returnVal != "[]" {
			//fmt.Println("error:", returnVal)
			responseChannel <- returnVal
			break
		}
	}
}

// ParseErrorResponseForTestSuccess parses the response of the Error channel.
// It prints the response to the response channel
func ParseErrorResponseForTestSuccess(message string, channel chan []byte, responseChannel chan string) {
	for {
		value, ok := <-channel
		if !ok {
			break
		}
		returnVal := string(value)
		if returnVal != "[]" {
			//fmt.Println("returnVal ", returnVal)
			if strings.Contains(returnVal, message) {
				responseChannel <- "passed"
			} else {
				responseChannel <- "failed"
			}
			break
		}
	}
}

// ParseResponseDummy is a methods that reads the response on the channel
// but does notthing on it.
func ParseResponseDummy(channel chan []byte) {
	for {
		value, ok := <-channel
		if !ok {
			break
		}
		returnVal := string(value)
		if returnVal != "[]" {
			//fmt.Println ("ParseSubscribeResponseDummy", returnVal)
			break
		}
	}
}

// ParseResponseDummy is a methods that reads the response on the channel
// but does notthing on it.
func ParseResponseDummyMessage(channel chan []byte, message string, responseChannel chan string) {
	for {
		value, ok := <-channel
		if !ok {
			break
		}
		returnVal := string(value)
		if returnVal != "[]" {
			//fmt.Println ("ParseSubscribeResponseDummy", returnVal)
			response := fmt.Sprintf("%s", value)
			if strings.Contains(response, "aborted") {
				continue
			}

			responseChannel <- returnVal
			break
		}
	}
}

func ExpectConnectedEvent(t *testing.T,
	channels, groups string, successChannel, errorChannel <-chan []byte) {

	var initialChannelsArray, initialGroupsArray []string

	if len(channels) > 0 {
		initialChannelsArray = strings.Split(channels, ",")
	}

	if len(groups) > 0 {
		initialGroupsArray = strings.Split(groups, ",")
	}

	waitForEventOnEveryChannel(t, initialChannelsArray, initialGroupsArray,
		"connected", successChannel, errorChannel)
}

func ExpectUnsubscribedEvent(t *testing.T,
	channels, groups string, successChannel, errorChannel <-chan []byte) {

	var initialChannelsArray, initialGroupsArray []string

	if len(channels) > 0 {
		initialChannelsArray = strings.Split(channels, ",")
	}

	if len(groups) > 0 {
		initialGroupsArray = strings.Split(groups, ",")
	}

	waitForEventOnEveryChannel(t, initialChannelsArray, initialGroupsArray,
		"unsubscribed", successChannel, errorChannel)
}

func waitForEventOnEveryChannel(t *testing.T, channels, groups []string,
	action string, successChannel, errorChannel <-chan []byte) {

	var triggeredChannels []string
	var triggeredGroups []string

	channel := make(chan bool)

	go func() {
		for {
			select {
			case event := <-successChannel:
				var ary []interface{}

				eventString := string(event)
				assert.Contains(t, eventString, action)

				err := json.Unmarshal(event, &ary)
				if err != nil {
					assert.Fail(t, err.Error())
				}

				if strings.Contains(eventString, "channel group") {
					triggeredGroups = append(triggeredGroups, ary[2].(string))
				} else if strings.Contains(eventString, "channel") {
					triggeredChannels = append(triggeredChannels, ary[2].(string))
				}
				if AssertStringSliceElementsEqual(triggeredChannels, channels) &&
					AssertStringSliceElementsEqual(triggeredGroups, groups) {
					channel <- true
					return
				}
			case err := <-errorChannel:
				assert.Fail(t, string(err))
				channel <- false
				return
			}
		}
	}()

	select {
	case <-channel:
	case <-timeouts(20):
		assert.Fail(t, fmt.Sprintf(
			"Timeout occured for %s event. Expected channels/groups: %s/%s. "+
				"Received channels/groups: %s/%s\n",
			action, channels, groups, triggeredChannels, triggeredGroups))
	}
}

func timeout() <-chan time.Time {
	return time.After(time.Second * time.Duration(testTimeout))
}

func timeouts(seconds int) <-chan time.Time {
	return time.After(time.Second * time.Duration(seconds))
}

func GenerateTwoRandomChannelStrings(length int) (channels1, channels2 string) {
	var channelsArray []string

	r := GenRandom()
	channelsMap := make(map[string]struct{})

	for len(channelsMap) < length*2 {
		channel := fmt.Sprintf("testChannel_sub_%d", r.Intn(99999))

		if _, found := channelsMap[channel]; !found {
			channelsMap[channel] = struct{}{}
		}
	}

	for channel := range channelsMap {
		channelsArray = append(channelsArray, channel)
	}

	return strings.Join(channelsArray[:length], ","), strings.Join(channelsArray[length:], ",")
}

func AssertStringSliceElementsEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}

	if len(first) == 0 && len(second) == 0 {
		return true
	}

	for _, f := range first {
		firstFound := false

		for _, s := range second {
			if f == s {
				firstFound = true
			}
		}

		if firstFound == false {
			return false
		}
	}

	return true
}

func RandomChannel() string {
	channel, _ := GenerateTwoRandomChannelStrings(1)
	return channel
}

func RandomChannels(length int) string {
	channel, _ := GenerateTwoRandomChannelStrings(length)
	return channel
}

func NewPubnubMatcher(skipFields []string) cassette.Matcher {
	matcher := &PubnubMatcher{}

	matcher.skipFields = skipFields

	return matcher
}

type PubnubMatcher struct {
	cassette.Matcher

	skipFields []string
}

func (m *PubnubMatcher) Match(interactions []*cassette.Interaction,
	r *http.Request) (*cassette.Interaction, error) {

interactionsLoop:
	for _, i := range interactions {
		if r.Method != i.Request.Method {
			continue
		}

		expectedURL, err := url.Parse(i.URL)
		if err != nil {
			continue
		}

		if expectedURL.Host != r.URL.Host {
			continue
		}

		if expectedURL.Path != r.URL.Path {
			continue
		}

		eQuery := expectedURL.Query()
		aQuery := r.URL.Query()

		for fKey, _ := range eQuery {
			if hasKey(fKey, m.skipFields) {
				continue
			}

			if aQuery[fKey] == nil || eQuery.Get(fKey) != aQuery.Get(fKey) {
				continue interactionsLoop
			}
		}

		return i, nil
	}

	return nil, cassette.InteractionNotFound
}

var pubnubMatcher cassette.Matcher = NewPubnubMatcher([]string{})

func hasKey(key string, list []string) bool {
	for _, v := range list {
		if v == key {
			return true
		}
	}

	return false
}

func LogErrors(errorsChannel <-chan []byte) {
	fmt.Printf("ERROR: %s", <-errorsChannel)
}
