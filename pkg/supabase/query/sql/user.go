package sql

import "fmt"

var GetUserByEmailQuery = `
select
  * from
  auth.users
  email = %s
limit 1
`

func GenerateGetUserQuery(email string) string {
	return fmt.Sprintf(GetUserByEmailQuery, Literal(email))
}

var UpdateUserRecoveryTokenQuery = `
update
  auth.users
set
  recovery_token = %s
  where
    email = %s
`

func GenerateUpdateUserRecoveryTokenQuery(token string, email string) string {
	return fmt.Sprintf(UpdateUserRecoveryTokenQuery, Literal(token), Literal(email))
}
