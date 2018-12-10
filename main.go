package main

import (
  "fmt"
  "context"
  "errors"
  "github.com/gofrs/uuid"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type MyEvent struct {
  Id string `json:"id"`
  Token string `json:"token"`
  Question string `json:"question"`
  Answer string `json:"answer"`
  Incorrect []string `json:"incorrect"`
}

type MyItem struct {
  id string
  token string
  question string
  answer string
  incorrect []string
}

func HandleRequest(ctx context.Context, event MyEvent) (string, error) {
  // Validate input
  // Is token valid? We should be able to find someone with this token in our customer DB

  // Are there three incorrect answers? Are all answers different?
  if len(event.Incorrect) != 3 {
    return "", errors.New("Need three incorrect answers")
  }
  if (event.Incorrect[0] == event.Incorrect[1]) || (event.Incorrect[1] == event.Incorrect[2]) || (event.Incorrect[0] == event.Incorrect[2]) {
    return "", errors.New("Duplicate answers")
  }
  for _, s := range event.Incorrect {
    if s == event.Answer {
      return "", errors.New("Duplicate answers")
    }
  }

  // Is the question less than 160 characters and each answer less than 80 characters?
  if len(event.Question) > 160 {
    return "", errors.New("Question length more than 160 characters")
  }
  if len(event.Answer) > 80 {
    return "", errors.New("Answer length more than 80 characters")
  }
  for _, s := range event.Incorrect {
    if len(s) > 80 {
      return "", errors.New("Answer length more than 80 characters")
    }
  }

  // OK, save to the DB
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-east-1")},
  )

  id, err := uuid.NewV4()
  event.Id = id.String()

  // Create DynamoDB client
  svc := dynamodb.New(sess)
  av, _ := dynamodbattribute.MarshalMap(event)
fmt.Println(av)
  input := &dynamodb.PutItemInput{
    TableName: aws.String("TriviaQuestions"),
    Item: av,
  }

  _, err = svc.PutItem(input)
  if err != nil {
    fmt.Println(err.Error())
    return "", err
  }
  return "OK", nil
}

func main() {
  lambda.Start(HandleRequest)
}
