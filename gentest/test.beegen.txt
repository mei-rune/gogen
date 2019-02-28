// Please don't edit this file!
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	beego "github.com/astaxie/beego"
	beecontext "github.com/astaxie/beego/context"
)

func httpCodeWith(err error) int {
	if herr, ok := err.(interface {
		HTTPCode() int
	}); ok {
		return herr.HTTPCode()
	}
	return http.StatusInternalServerError
}

func InitStringSvc(mux *beego.Namespace, svc StringSvc) {
	mux.Get("/echo", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")

		result := svc.Echo(a)
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/echo", func(ctx *beecontext.Context) {

		result, err := svc.EchoBody(ctx.Request.Body)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")
		var b = ctx.Input.Query("b")

		result, err := svc.Concat(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat1", func(ctx *beecontext.Context) {
		var a *string
		if s := ctx.Input.Query("a"); s != "" {
			a = &s
		}
		var b *string
		if s := ctx.Input.Query("b"); s != "" {
			b = &s
		}

		result, err := svc.Concat1(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat2/:a/:b", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b = ctx.Input.Param("b")

		result, err := svc.Concat2(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat3/:a/:b", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b = ctx.Input.Param("b")

		result, err := svc.Concat3(&a, &b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/sub", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")
		var start int64
		if s := ctx.Input.Query("start"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", start, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", start, s, err).Error())
				return
			}
			start = v64
		}

		result, err := svc.Sub(a, start)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Post("/save/:a", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b string
		if err := json.Unmarshal(ctx.Input.CopyBody(4*1024), &b); err != nil {
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err).Error())
			return
		}

		result, err := svc.Save(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusCreated)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Post("/save2/:a", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b string
		if err := json.Unmarshal(ctx.Input.CopyBody(4*1024), &b); err != nil {
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err).Error())
			return
		}

		result, err := svc.Save2(&a, &b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusCreated)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add/:a/:b", func(ctx *beecontext.Context) {
		var a int
		if v64, err := strconv.ParseInt(ctx.Input.Param("a"), 10, 64); err != nil {
			s := ctx.Input.Param("a")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
			return
		} else {
			a = int(v64)
		}
		var b int
		if v64, err := strconv.ParseInt(ctx.Input.Param("b"), 10, 64); err != nil {
			s := ctx.Input.Param("b")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
			return
		} else {
			b = int(v64)
		}

		result, err := svc.Add(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add2/:a/:b", func(ctx *beecontext.Context) {
		var a *int
		if v64, err := strconv.ParseInt(ctx.Input.Param("a"), 10, 64); err != nil {
			s := ctx.Input.Param("a")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
			return
		} else {
			a = new(int)
			*a = int(v64)
		}
		var b *int
		if v64, err := strconv.ParseInt(ctx.Input.Param("b"), 10, 64); err != nil {
			s := ctx.Input.Param("b")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
			return
		} else {
			b = new(int)
			*b = int(v64)
		}

		result, err := svc.Add2(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add3", func(ctx *beecontext.Context) {
		var a *int
		if s := ctx.Input.Query("a"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
				return
			}
			a = new(int)
			*a = int(v64)
		}
		var b *int
		if s := ctx.Input.Query("b"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
				return
			}
			b = new(int)
			*b = int(v64)
		}

		result, err := svc.Add3(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	// Misc: annotation is missing
}

func InitStringSvcImpl(mux *beego.Namespace, svc *StringSvcImpl) {
	mux.Get("/echo", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")

		result := svc.Echo(a)
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/echo_body", func(ctx *beecontext.Context) {

		result, err := svc.EchoBody(ctx.Request.Body)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")
		var b = ctx.Input.Query("b")

		result, err := svc.Concat(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat1", func(ctx *beecontext.Context) {
		var a *string
		if s := ctx.Input.Query("a"); s != "" {
			a = &s
		}
		var b *string
		if s := ctx.Input.Query("b"); s != "" {
			b = &s
		}

		result, err := svc.Concat1(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat2/:a/:b", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b = ctx.Input.Param("b")

		result, err := svc.Concat2(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat3/:a/:b", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b = ctx.Input.Param("b")

		result, err := svc.Concat3(&a, &b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/sub", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")
		var start int64
		if s := ctx.Input.Query("start"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", start, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", start, s, err).Error())
				return
			}
			start = v64
		}

		result, err := svc.Sub(a, start)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Post("/save/:a", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b string
		if err := json.Unmarshal(ctx.Input.CopyBody(4*1024), &b); err != nil {
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err).Error())
			return
		}

		result, err := svc.Save(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusCreated)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Post("/save2/:a", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b string
		if err := json.Unmarshal(ctx.Input.CopyBody(4*1024), &b); err != nil {
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err).Error())
			return
		}

		result, err := svc.Save2(&a, &b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusCreated)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add/:a/:b", func(ctx *beecontext.Context) {
		var a int
		if v64, err := strconv.ParseInt(ctx.Input.Param("a"), 10, 64); err != nil {
			s := ctx.Input.Param("a")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
			return
		} else {
			a = int(v64)
		}
		var b int
		if v64, err := strconv.ParseInt(ctx.Input.Param("b"), 10, 64); err != nil {
			s := ctx.Input.Param("b")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
			return
		} else {
			b = int(v64)
		}

		result, err := svc.Add(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add2/:a/:b", func(ctx *beecontext.Context) {
		var a *int
		if v64, err := strconv.ParseInt(ctx.Input.Param("a"), 10, 64); err != nil {
			s := ctx.Input.Param("a")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
			return
		} else {
			a = new(int)
			*a = int(v64)
		}
		var b *int
		if v64, err := strconv.ParseInt(ctx.Input.Param("b"), 10, 64); err != nil {
			s := ctx.Input.Param("b")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
			return
		} else {
			b = new(int)
			*b = int(v64)
		}

		result, err := svc.Add2(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add3", func(ctx *beecontext.Context) {
		var a *int
		if s := ctx.Input.Query("a"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
				return
			}
			a = new(int)
			*a = int(v64)
		}
		var b *int
		if s := ctx.Input.Query("b"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
				return
			}
			b = new(int)
			*b = int(v64)
		}

		result, err := svc.Add3(a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	// Misc: annotation is missing
}

func InitStringSvcWithContext(mux *beego.Namespace, svc *StringSvcWithContext) {
	mux.Get("/echo", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")

		result := svc.Echo(ctx.Request.Context(), a)
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/echo", func(ctx *beecontext.Context) {

		result, err := svc.EchoBody(ctx.Request.Context(), ctx.Request.Body)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")
		var b = ctx.Input.Query("b")

		result, err := svc.Concat(ctx.Request.Context(), a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat1", func(ctx *beecontext.Context) {
		var a *string
		if s := ctx.Input.Query("a"); s != "" {
			a = &s
		}
		var b *string
		if s := ctx.Input.Query("b"); s != "" {
			b = &s
		}

		result, err := svc.Concat1(ctx.Request.Context(), a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat2/:a/:b", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b = ctx.Input.Param("b")

		result, err := svc.Concat2(ctx.Request.Context(), a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/concat3/:a/:b", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b = ctx.Input.Param("b")

		result, err := svc.Concat3(ctx.Request.Context(), &a, &b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/sub", func(ctx *beecontext.Context) {
		var a = ctx.Input.Query("a")
		var start int64
		if s := ctx.Input.Query("start"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", start, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", start, s, err).Error())
				return
			}
			start = v64
		}

		result, err := svc.Sub(ctx.Request.Context(), a, start)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Post("/save/:a", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b string
		if err := json.Unmarshal(ctx.Input.CopyBody(4*1024), &b); err != nil {
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err).Error())
			return
		}

		result, err := svc.Save(ctx.Request.Context(), a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusCreated)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Post("/save2/:a", func(ctx *beecontext.Context) {
		var a = ctx.Input.Param("a")
		var b string
		if err := json.Unmarshal(ctx.Input.CopyBody(4*1024), &b); err != nil {
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, "<no value>", err).Error())
			return
		}

		result, err := svc.Save2(ctx.Request.Context(), &a, &b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusCreated)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add/:a/:b", func(ctx *beecontext.Context) {
		var a int
		if v64, err := strconv.ParseInt(ctx.Input.Param("a"), 10, 64); err != nil {
			s := ctx.Input.Param("a")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
			return
		} else {
			a = int(v64)
		}
		var b int
		if v64, err := strconv.ParseInt(ctx.Input.Param("b"), 10, 64); err != nil {
			s := ctx.Input.Param("b")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
			return
		} else {
			b = int(v64)
		}

		result, err := svc.Add(ctx.Request.Context(), a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add2/:a/:b", func(ctx *beecontext.Context) {
		var a *int
		if v64, err := strconv.ParseInt(ctx.Input.Param("a"), 10, 64); err != nil {
			s := ctx.Input.Param("a")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
			return
		} else {
			a = new(int)
			*a = int(v64)
		}
		var b *int
		if v64, err := strconv.ParseInt(ctx.Input.Param("b"), 10, 64); err != nil {
			s := ctx.Input.Param("b")
			ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
			ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
			return
		} else {
			b = new(int)
			*b = int(v64)
		}

		result, err := svc.Add2(ctx.Request.Context(), a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	mux.Get("/add3", func(ctx *beecontext.Context) {
		var a *int
		if s := ctx.Input.Query("a"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", a, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
				return
			}
			a = new(int)
			*a = int(v64)
		}
		var b *int
		if s := ctx.Input.Query("b"); s != "" {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.Output.SetStatus(httpCodeWith(fmt.Errorf("argument %q is invalid - %q", b, s, err)))
				ctx.WriteString(fmt.Errorf("argument %q is invalid - %q", b, s, err).Error())
				return
			}
			b = new(int)
			*b = int(v64)
		}

		result, err := svc.Add3(ctx.Request.Context(), a, b)
		if err != nil {
			ctx.Output.SetStatus(httpCodeWith(err))
			ctx.WriteString(err.Error())
			return
		}
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(result, false, false)
		return
	})
	// Misc: annotation is missing
}
