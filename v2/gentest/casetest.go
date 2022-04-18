package main

import (
	"database/sql"
)

// @http.Client(name="TestClient", ref="true")
type CaseSvc interface {

	// @Summary test by name
	// @Description test by int64 ID
	// @ID TestCase1
	// @Accept  json
	// @Produce  json
	// @Param   name      path   string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name/{name} [get]
	TestCase1(name string) error

	// @Summary test by name
	// @Description test by int64 ID
	// @ID TestCase2_1
	// @Accept  json
	// @Produce  json
	// @Param   name      query   string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name [get]
	TestCase2_1(name string) error

	// @Summary test by names
	// @Description test by int64 ID
	// @ID TestCase2_2
	// @Accept  json
	// @Produce  json
	// @Param   name      query   []string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_names [get]
	TestCase2_2(name []string) error

	// @Summary test by int64 ID
	// @Description test by int64 ID
	// @ID TestCase3_1
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int64)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase3_1(id int64) error

	// @Summary test by int32 ID
	// @Description test by int64 ID
	// @ID TestCase3_2
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int32)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase3_2(id int32) error

	// @Summary test by int ID
	// @Description test by int64 ID
	// @ID TestCase3_3
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase3_3(id int) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以不好测试， iris 才支持)
	// @Description test by int64 ID
	// @ID TestCase4
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase4(id int) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以正常测试， iris 无法测试)
	// @Description test by int64 ID
	// @ID TestCase5
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase5_1(id int64) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以正常测试， iris 无法测试)
	// @Description test by int64 ID
	// @ID TestCase5
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase5_2(id int32) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以正常测试， iris 无法测试)
	// @Description test by int64 ID
	// @ID TestCase5
	// @Accept  json
	// @Produce  json
	// @Param   idlist      query   []int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase5_3(idlist []int64) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以无法测试， iris 可以测试)
	// @Description test by int64 ID
	// @ID TestCase6
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase6(id int64) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以测试， iris 不能正确测试)
	// @Description test by int64 ID
	// @ID TestCase7_1
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase7_1(id sql.NullInt64) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以测试， iris 不能正确测试)
	// @Description test by int32 ID
	// @ID TestCase7_2
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase7_2(id sql.NullInt32) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以不能正确地测试， iris 能正确测试)
	// @Description test by int64 ID
	// @ID TestCase8
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase8(id sql.NullInt64) error

	// @Summary test by int ID
	// @Description test by int64 ID
	// @ID TestCase9
	// @Accept  json
	// @Produce  json
	// @Param   id      path   string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase9(id *string) error

	// @Summary test by id
	// @Description test by string ID
	// @ID TestCase10
	// @Accept  json
	// @Produce  json
	// @Param   id      query   string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name [get]
	TestCase10(id *string) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以不能正确地测试， iris 能正确测试)
	// @Description test by int64 ID
	// @ID TestCase12
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/{id} [get]
	TestCase12(id *int) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以正确地测试， iris 不能正确测试)
	// @Description test by int64 ID
	// @ID TestCase13
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/{id} [get]
	TestCase13(id *int) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Query，所以可以正确地测试， iris 不能正确测试)
	// @Description test by int64 ID
	// @ID TestCase14_1
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase14_1(id *int) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Query，所以可以正确地测试， iris 不能正确测试)
	// @Description test by int64 ID
	// @ID TestCase14_2
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name [get]
	TestCase14_2(id *int32) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Query，所以不能正确地测试， iris 可以正确测试)
	// @Description test by int64 ID
	// @ID TestCase14_3
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name [get]
	TestCase14_3(id *int) error

	// Misc() string
}