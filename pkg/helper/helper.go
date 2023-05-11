package helper

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"strconv"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	digits   = "0123456789"
	specials = "~=+%^*/()[]{}/!@#$?|"
	all      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits
)

func ReplaceQueryParams(namedQuery string, params map[string]interface{}) (string, []interface{}) {
	var (
		i    int = 1
		args []interface{}
	)

	for k, v := range params {
		if k != "" && strings.Contains(namedQuery, ":"+k) {
			namedQuery = strings.ReplaceAll(namedQuery, ":"+k, "$"+strconv.Itoa(i))
			args = append(args, v)
			i++
		}
	}

	return namedQuery, args
}

func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

func ConvertMapToStruct(inputMap map[string]interface{}) (*structpb.Struct, error) {
	log.Printf("---INFO->ConvertMapToStruct---> %+v", inputMap)

	marshledInputMap, err := json.Marshal(inputMap)
	outputStruct := &structpb.Struct{}
	if err != nil {
		return outputStruct, err
	}
	err = protojson.Unmarshal(marshledInputMap, outputStruct)

	return outputStruct, err
}

func ConvertRequestToSturct(inputRequest interface{}) (*structpb.Struct, error) {
	marshelledInputInterface, err := json.Marshal(inputRequest)
	outputStruct := &structpb.Struct{}
	if err != nil {
		return outputStruct, err
	}
	err = protojson.Unmarshal(marshelledInputInterface, outputStruct)
	return outputStruct, err
}

func ConvertStructToResponse(inputStruct *structpb.Struct) (map[string]interface{}, error) {
	marshelledInputStruct, err := protojson.Marshal(inputStruct)
	outputMap := make(map[string]interface{}, 0)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(marshelledInputStruct, &outputMap)
	return outputMap, err
}

func ConverPhoneNumberToMongoPhoneFormat(input string) string {
	//input +998995677777
	input = input[4:]

	// input  = 995677777
	changedEl := input[:2]

	input = "(" + changedEl + ") " + input[2:5] + "-" + input[5:7] + "-" + input[7:]

	// input = (99) 567-77-77
	return input
}

func GenerateSecretKey(length int) string {
	buf := make([]byte, length)

	for i := 0; i < length; i++ {
		buf[i] = all[cryptoRandSecure(int64(len(all)))]
	}

	return string(buf)
}

func cryptoRandSecure(max int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		log.Println(err)
	}
	return nBig.Int64()
}
