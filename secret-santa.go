package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/smtp"
	"os"
	"sort"
	"strconv"
	"time"
)

func main() {
	configFileName := "config.json"

	if len(os.Args) < 2 {
		fmt.Println("Please provide a configuration json file (config.json)")
		fmt.Println("example: secret-santa config.json")
		return
	}

	configFileName = os.Args[1]

	maybeConfig, configErr := loadConfig(configFileName)

	if configErr != nil {
		panic(configErr)
	}

	globalConfig = *maybeConfig

	addresses, addressesErr := loadData(globalConfig.DataFile)
	if addressesErr != nil {
		panic(addressesErr)
	}

	assignments, assignmentsErr := generateSecretSantaAssignments(addresses)
	if assignmentsErr != nil {
		panic(assignmentsErr)
	}

	for _, assignment := range assignments {
		err := sendSecretSantaEmail(assignment, globalConfig.HTMLTemplateFile)
		if err != nil {
			panic(err)
		}
	}
}

var globalConfig Config

func loadConfig(filename string) (*Config, error) {

	fileData, fileErr := os.ReadFile(filename)

	if fileErr != nil {
		return nil, fmt.Errorf(`Error reading confiuration File "%v", Source Error: %v`, filename, fileErr)
	}

	var config Config

	parseError := json.Unmarshal(fileData, &config)

	if parseError != nil {
		return nil, fmt.Errorf(`Error parsing confiuration File "%v", Source Error: %v`, filename, parseError)
	}

	return &config, nil
}

func printlnDebug(a ...any) (n int, err error) {
	if globalConfig.EnableDebugMessages {
		return fmt.Println(a)
	}
	return 0, nil
}

func loadData(fileName string) ([]Address, error) {
	fileData, fileDataErr := os.ReadFile(fileName)

	if fileDataErr != nil {
		return nil, fmt.Errorf("Error loading file %v", fileDataErr)
	}

	var addresses []Address

	unmarshalErr := json.Unmarshal(fileData, &addresses)

	if unmarshalErr != nil {
		return nil, fmt.Errorf("Error parsing file %v", unmarshalErr)
	}

	err := validateData(addresses)

	if err != nil {
		panic(err)
	}

	return addresses, nil
}

func validateData(addresses []Address) error {
	residentMap := createResidentMap(addresses)
	idMap := make(map[string]bool, len(residentMap))

	// validate the addresses
	for _, address := range addresses {
		if len(address.Address) <= 0 {
			return errors.New(fmt.Sprintf("Data File Error: Missing \"address\"\nJSON:\n%v", toJson(address)))
		}
		if len(address.City) <= 0 {
			return errors.New(fmt.Sprintf("Data File Error: Missing \"city\"\nJSON:\n%v", toJson(address)))
		}
		if len(address.State) <= 0 {
			return errors.New(fmt.Sprintf("Data File Error: Missing \"state\"\nJSON:\n%v", toJson(address)))
		}
		if len(address.Zipcode) <= 0 {
			return errors.New(fmt.Sprintf("Data File Error: Missing \"zipcode\"\nJSON:\n%v", toJson(address)))
		}

		// validate the people
		for _, person := range address.Residents {
			if len(person.Id) <= 0 {
				return errors.New(fmt.Sprintf("Data File Error: Resident missing \"id\" under Address \"%v\"\nJSON:\n%v", address.Address, toJson(person)))
			}

			_, exists := idMap[person.Id]
			if exists {
				return errors.New(fmt.Sprintf("Data File Error: Resident \"id\" value \"%v\" must be unique\nJSON:\n%v", person.Id, toJson(person)))
			}
			idMap[person.Id] = true

			if len(person.FirstName) <= 0 {
				return errors.New(fmt.Sprintf("Data File Error: Resident missing \"first_name\" under Address \"%v\"\nJSON:\n%v", address.Address, toJson(person)))
			}
			if len(person.LastName) <= 0 {
				return errors.New(fmt.Sprintf("Data File Error: Resident missing \"last_name\" under Address \"%v\"\nJSON:\n%v", address.Address, toJson(person)))
			}
			if len(person.PublicName) <= 0 {
				return errors.New(fmt.Sprintf("Data File Error: Resident missing \"public_name\" under Address \"%v\"\nJSON:\n%v", address.Address, toJson(person)))
			}
			if len(person.Email) <= 0 {
				return errors.New(fmt.Sprintf("Data File Error: Resident missing \"email\" under Address \"%v\"\nJSON:\n%v", address.Address, toJson(person)))
			}

			if len(person.ActorId) > 0 {

				if person.ActorId == person.Id {
					return errors.New(fmt.Sprintf("Data File Error: \"actor_id\" \"%v\" must be different from \"id\"\nJSON:\n%v", person.ActorId, toJson(person)))
				}
				_, exists := residentMap[person.ActorId]

				if !exists {
					return errors.New(fmt.Sprintf("Data File Error: \"actor_id\" \"%v\" must be an existing \"id\" value from a different person\nJSON:\n%v", person.ActorId, toJson(person)))
				}
			}
		}
	}
	return nil
}

func toJson(a any) string {
	bytes, _ := json.MarshalIndent(a, "", "\t")
	return string(bytes)
}

func createResidentMap(addresses []Address) map[string]Resident {
	recipients := make(map[string]Resident, 10)
	for _, address := range addresses {
		for _, resident := range address.Residents {
			recipients[resident.Id] = Resident{
				Person:  resident,
				Address: address,
			}
		}
	}

	return recipients
}

func generateSecretSantaAssignments(addresses []Address) ([]SecretSantaAssignment, error) {

	recipients := createResidentMap(addresses)

	// randomize!
	rand.Seed(time.Now().UnixMilli())

	run := true
	count := 0
	for run && count < 1000 {
		assignments, err := attemptSecretSantaAssignments(recipients)

		if err == nil {
			return assignments, nil
		}

		printlnDebug("Reattempting assignment...")

		count++

		if count == 1000 {
			if !globalConfig.Rules.AllowSameResidenceExchange {
				// allowing same residence exchanges are more forgiving, so this hopefully should never get in here.
				return nil, fmt.Errorf(`Could not assign secret santas..`)
			}
			return nil, fmt.Errorf(`Could not assign secret santas. Try setting "allow_same_residence_exchange" to true.`)
		}
	}

	return nil, nil
}

func randomSequenceFromMapValues[K comparable, V any](input map[K]V) []V {
	sequence := make([]randElement[V], len(input))

	for _, element := range input {
		sequence = append(sequence, randElement[V]{
			Order:   rand.Int(),
			Element: element,
		})
	}

	sort.Slice(sequence, func(i, j int) bool {
		return sequence[i].Order > sequence[j].Order
	})

	result := make([]V, len(input))
	for index := 0; index < len(input); index++ {
		result[index] = sequence[index].Element
	}

	return result
}

type randElement[T any] struct {
	Order   int
	Element T
}

func attemptSecretSantaAssignments(residents map[string]Resident) ([]SecretSantaAssignment, error) {

	residentsCopy := make(map[string]Resident)
	for k, v := range residents {
		residentsCopy[k] = v
	}

	residentSequence := randomSequenceFromMapValues(residentsCopy)

	assignments := make([]SecretSantaAssignment, 0)

	for _, from := range residentSequence {
		success := false

		if len(residentsCopy) == 0 {
			break
		}

		recipientList := randomSequenceFromMapValues(residentsCopy)

		for _, to := range recipientList {
			if from.Person.Id == to.Person.Id {
				continue
			}

			_, exists := residentsCopy[to.Person.Id]

			if !exists {
				printlnDebug(fmt.Sprintf("Person %v does not exist in the RecipientsCopy", to.Person.Id))
				continue
			}

			if !globalConfig.Rules.AllowSameResidenceExchange && from.Address.Address == to.Address.Address {
				printlnDebug(fmt.Sprintf("Person %v and %v have the same address", from.Person.Id, to.Person.Id))
				continue
			}

			assignment := SecretSantaAssignment{
				From:  from,
				To:    to,
				Actor: from,
			}

			if len(from.Person.ActorId) > 0 {
				actor := residents[from.Person.ActorId]
				assignment.Actor = actor
			}

			assignments = append(assignments, assignment)
			delete(residentsCopy, to.Person.Id)
			success = true
			break
		}

		if !success {
			return nil, errors.New("Unable to find a pairing")
		}
	}

	return assignments, nil
}

func sendSecretSantaEmail(assignment SecretSantaAssignment, emailTemplate string) error {

	templateHtml, templateHtmlErr := template.ParseFiles(emailTemplate)

	if templateHtmlErr != nil {
		return fmt.Errorf("Unable to read email templateHtml %v", templateHtmlErr)
	}

	// process templates
	buf := new(bytes.Buffer)

	executeErr := templateHtml.Execute(buf, assignment)

	if executeErr != nil {
		return fmt.Errorf("Unable to templatize email %v", executeErr)
	}

	emailRecipient := assignment.Actor.Person

	printlnDebug(fmt.Sprintf("Emailed To: %s, For %s → %s", emailRecipient.Email, assignment.From.Person.Id, assignment.To.Person.Id))

	if globalConfig.WriteHtmlFiles {
		folder := "output_html"
		if _, err := os.Stat(folder); os.IsNotExist(err) {
			os.Mkdir(folder, 0755)
		}

		fileName := fmt.Sprintf("%v/%v.html", folder, assignment.From.Person.Id)
		fmt.Println("Writing", fileName)

		os.WriteFile(fileName, buf.Bytes(), 0755)
	}

	emailErr := sendHtmlEmail(emailRecipient, string(buf.Bytes()))

	if emailErr != nil {
		return fmt.Errorf("Unable to send email %v", emailErr)
	}

	return nil
}

func sendHtmlEmail(to Person, htmlBody string) error {
	if !globalConfig.SMTP.Enabled {
		return nil
	}

	host := globalConfig.SMTP.Host + ":" + strconv.Itoa(globalConfig.SMTP.Port)

	auth := smtp.PlainAuth("", globalConfig.SMTP.UserName, globalConfig.SMTP.Password, globalConfig.SMTP.Host)

	senderEmail := "Secret Santa <santa@codereactor.com>"
	recipientEmail := fmt.Sprintf("%s <%s>", to.PublicName, to.Email)
	messageHeaders := make(map[string]string, 3)
	messageHeaders["Subject"] = globalConfig.SMTP.Subject
	messageHeaders["From"] = senderEmail
	messageHeaders["To"] = recipientEmail

	otherHeaders := make(map[string]string, 2)
	otherHeaders["MIME-version"] = "1.0"
	otherHeaders["Content-Type"] = `text/html; charset="UTF-8"`

	body := ""

	for key, value := range messageHeaders {
		body = fmt.Sprintf("%v%v: %v\n", body, key, value)
	}

	for key, value := range otherHeaders {
		body = fmt.Sprintf("%v%v: %v;\n", body, key, value)
	}

	body = body + htmlBody
	msg := []byte(body)

	if err := smtp.SendMail(host, auth, senderEmail, []string{recipientEmail}, msg); err != nil {
		return err
	}
	return nil
}

type Config struct {
	DataFile            string      `json:"data_file"`
	HTMLTemplateFile    string      `json:"html_template_file"`
	EnableDebugMessages bool        `json:"enable_debug_messages"`
	WriteHtmlFiles      bool        `json:"write_html_files"`
	Rules               RulesConfig `json:"rules"`
	SMTP                SMTPConfig  `json:"smtp"`
}

type SMTPConfig struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	UserName    string `json:"user_name"`
	Password    string `json:"password"`
	SenderEmail string `json:"sender_email"`
	Subject     string `json:"subject"`
}

type RulesConfig struct {
	AllowSameResidenceExchange bool `json:"allow_same_residence_exchange"`
}

type Address struct {
	Address   string   `json:"address"`
	City      string   `json:"city"`
	State     string   `json:"state"`
	Zipcode   string   `json:"zipcode"`
	Residents []Person `json:"residents"`
}

type Person struct {
	Id         string `json:"id"`
	ActorId    string `json:"actor_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PublicName string `json:"public_name"`
	Email      string `json:"email"`
}

type Resident struct {
	Person  Person
	Address Address
}

type SecretSantaAssignment struct {
	From  Resident
	To    Resident
	Actor Resident
}
