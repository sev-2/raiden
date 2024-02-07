package raiden_test

// // setup
// type LoginHistory struct {
// 	Id     uuid.UUID `json:"id"`
// 	UserId uuid.UUID `json:"user_id"`
// }

// type RolePermissions struct {
// 	Id     uuid.UUID `json:"id"`
// 	RoleId uuid.UUID `json:"role_id"`
// 	Name   string    `json:"name"`
// }

// type Role struct {
// 	Id          uuid.UUID          `json:"id"`
// 	Name        string             `json:"name"`
// 	Permissions []*RolePermissions `json:"role_permissions"`
// }

// type User struct {
// 	Id           uuid.UUID       `json:"id"`
// 	RoleId       int             `json:"role_id"`
// 	Name         string          `json:"string"`
// 	Level        int             `json:"level"`
// 	Type         string          `json:"type"`
// 	Role         Role            `json:"role" sourceColumn:"id" targetColumn:"role_id"`
// 	LoginHistory []*LoginHistory `json:"permission" sourceColumn:"user_id" targetColumn:"id"`
// }

// type Param struct {
// 	Page   int
// 	Limit  int
// 	Name   *string
// 	Levels []int
// 	Type   *string `default:"regular"`
// }

// func TestGenerateRpcQuery(t *testing.T) {
// 	param := Param{
// 		Name:   nil,
// 		Levels: []int{1},
// 	}

// 	rpc := raiden.Rpc{}
// 	rpc.SetName("test_function").SetModel(User{}, "u").SetParamStruct(param).SetQuery(`
// 		BEGIN
// 			RETURN QUERY
// 			SELECT * FROM :u u
// 			LEFT JOIN :u.Role r on u.role_id = r.id
// 			LEFT JOIN :u.LoginHistory lh on lh.user_id = u.id
// 			LEFT JOIN :u.Role.Permissions rp on rp.role_id = r.id
// 			OFFSET (:page - 1) * :limit LIMIT :limit;
// 		END
// 	`)

// 	err := rpc.BuildQuery()
// 	assert.NoError(t, err)

// 	expectedQuery := "create or replace function public.test_function(in_page integer,in_limit integer,in_name text default null,in_levels integer[],in_type text default 'regular') returns setof user language plpgsql security definer as $function$ begin return query select * from user u left join role r on u.role_id = r.id left join login_history lh on lh.user_id = u.id left join role rp on rp.role_id = r.id offset (in_page - 1) * in_limit limit in_limit; end $function$"
// 	assert.Equal(t, expectedQuery, rpc.CompleteQuery)
// 	assert.Equal(t, utils.HashString(expectedQuery), rpc.Hash)
// }
