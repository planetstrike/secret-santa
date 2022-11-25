# Secret Santa
Randomized gift exchange emailer.

## Configuration JSON

| Field                                 | Type    | Description                                                         | Example                                   |
|---------------------------------------|---------|---------------------------------------------------------------------|-------------------------------------------|
| `data_file`                           | string  | the file name that contains the addresses and people                | `"example_data.json"`                     |
| `html_template_file`                  | string  | html email template file name.                                      | `"example_template.html"`                 |
| `enable_debug_messages`               | boolean | writes debug messages to the console                                | `true`                                    |
| `write_html_files`                    | boolean | writes processed template html files to the `output_html` directory | `true`                                    |
| `rules.allow_same_residence_exchange` | boolean | allow gift exchanges to occur in the same household                 | `false`                                   |
| `smtp.enabled`                        | boolean | enables sending of the emails                                       | `true`                                    |
| `smtp.host`                           | string  | SMTP server host                                                    | `"in-v3.mailjet.com"`                     |
| `smtp.port`                           | int     | SMTP server port                                                    | `587`                                     |
| `smtp.user_name`                      | string  | SMTP plain auth user name                                           | `"generatedAPIKey1234"`                   |
| `smtp.password`                       | string  | SMTP plain auth password                                            | `"generatedAPIPassword5555"`              |
| `smtp.sender_email`                   | string  | Email `From` field on emails sent                                   | `"Secret Santa <santa@customdomain.com>"` |
| `smtp.subject`                        | string  | Email `Subject` field on emails sent                                | `"2022 Secret Santa!"`                    |

### Example
```json
{
  "data_file": "example_data.json",
  "html_template_file": "example_template.html",
  "enable_debug_messages": true,
  "write_html_files": true,
  "smtp": {
    "enabled" : false,
    "host" : "in-v3.mailjet.com",
    "port" : 587,
    "user_name" : "your_username",
    "password" : "secure_password",
    "sender_email": "Secret Santa <secretsanta@host.com>",
    "subject": "Secret Santa!"
  }
}
```

## Data File Format
| Field                   | Type   | Description                                                                                                                                | Example            |
|-------------------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------|--------------------|
| `address`               | string | mailing street address                                                                                                                     | `"45 Beach Ave."`  |
| `city`                  | string | mailing city                                                                                                                               | `"Atlanta"`        |
| `state`                 | string | mailing state                                                                                                                              | `"GA"`             |
| `zipcode`               | string | mailing state                                                                                                                              | `"30303"`          |
| `residents.id`          | string | a unique string used to identify people.                                                                                                   | `"jonathan_wise"`  |
| `residents.first_name`  | string | participant's first name                                                                                                                   | `"Jon"`            |
| `residents.last_name`   | string | participant's last name                                                                                                                    | `"Wise"`           |
| `residents.public_name` | string | when the official name used to send mail is different from what a person is normally called                                                | `"Jonathan Wise"`  |
| `residents.email`       | string | participant's email.                                                                                                                       | `"jigsaw@aol.com"` |
| `residents.actor_id`    | string | in cases where someone needs assistance, the `actor_id` is the `id` of another person who will receive the email on this person's behalf.  | `"elizabeth_wise"`   |

### Example
```json
[
  {
    "address" : "45 Beech Ave.",
    "city": "Atlanta",
    "state": "GA",
    "zipcode": "30303",
    "residents": [
      {
        "id": "jonathan_wise",
        "first_name": "Jon",
        "last_name": "Wise",
        "public_name": "Jonathan Wise",
        "email": "jigsaw@aol.com",
        "actor_id": "elizabeth_wise"
      },
      {
        "id": "elizabeth_wise",
        "first_name":"Liz",
        "last_name": "Wise",
        "public_name": "Elizabeth Wise",
        "email": "ournews@yahoo.com"
      }
    ]
  },
  {
    "address": "17 Blue Spring St",
    "city": "Harrison Township",
    "state": "MI",
    "zipcode": "48045",
    "residents": [
      {
        "id": "brett_curry",
        "first_name": "Brett",
        "last_name": "Curry",
        "public_name": "Brett Curry",
        "email": "yankees2232@outlook.com"
      },
      {
        "id": "caroline_curry",
        "first_name": "Caroline",
        "last_name": "Curry",
        "public_name": "Caroline Curry",
        "email": "seamonkey22@gmail.com"
      },
      {
        "id": "matias_curry",
        "first_name": "Matt",
        "last_name": "Curry",
        "public_name": "Matias Curry",
        "email": "bulbasaur123@gmail.com"
      }
    ]
  }
]
```


## HTML Templating
The templating payload is comprised of three main values. The primary payload is the `SecretSantaAssignment` type which is comprised of
three values. To insert a value into the HTML page, use `{{.?}}` to insert values. For example `{{.To.Person.FirstName}}`.

### SecretSantaAssignment
| Name    | Type       | Description                                                      |
|---------|------------|------------------------------------------------------------------|
| `To`    | `Resident` | Recipient of the secret santa present                            |
| `From`  | `Resident` | Secret Santa                                                     |
| `Actor` | `Resident` | Secret Santa's actor who is acting on behalf of the Secret Santa |

### Resident
| Name      | Type       | Description                                         |
|-----------|------------|-----------------------------------------------------|
| `Person`  | `Person`   | Person information                                  |
| `Address` | `Address`  | Person mailing address                              |

### Person
| Name         | Type   | Description      |
|--------------|--------|------------------|
| `Id`         | string | Person unique id |
| `ActorId`    | string | Actor id         |
| `FirstName`  | string | First Name       |
| `LastName`   | string | Last Name        |
| `PublicName` | string | Public Name      |
| `Email`      | string | Email address    |

### `Address`
| Field                      | Type   | Description            |
|----------------------------|--------|------------------------|
| `Address`                  | string | mailing street address |
| `City`                     | string | mailing city           |
| `State`                    | string | mailing state          |
| `Zipcode`                  | string | mailing state          |

### Examples

```html
<html>
<body>
Hello {{.From.Person.FirstName}}!

You are {{.To.Person.FirstName}}'s Secret Santa!

Please send your gift to
{{.To.Person.PublicName}}
{{.To.Address.Address}}
{{.To.Address.City}}, {{.To.Address.State}} {{.To.Address.Zipcode}}
</body>
</html>
```