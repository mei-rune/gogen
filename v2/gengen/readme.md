
		情况1
		var a = ctx.Param("a")

		情况2
		var a = ctx.Query("a")
		var names = ctx.GetQueryArray("names")
		var a = ctx.GetIntQueryWithDefault("a", 3)
		


		情况3
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, ctx.Param("id"), err).Error())
			return
		}
		idValue, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, ctx.Param("id"), err).Error())
			return
		}
		id := int32(idValue)


		情况4
		id, err := ctx.GetInt64Param("id")
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, ctx.Param("id"), err).Error())
			return
		}



		情况5
		var id int64
		if s := ctx.Query("id"); s != "" {
			idValue, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, s, err).Error())
				return
			}
			id = idValue
		}
		var id int32
		if s := ctx.Query("id"); s != "" {
			idValue, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, s, err).Error())
				return
			}
			id = int32(idValue)
		}
		var idArray []int64
		if ss := ctx.QueryArray("id_list"); len(ss) != 0 {
			idValue, err := toInts(ss)
			if err != nil {
				ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, s, err).Error())
				return
			}
			idArray = idValue
		}


		情况6
		id, err := ctx.GetInt64Query("id")
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, ctx.Param("id"), err).Error())
			return
		}



		情况7
		var id sql.NullInt64
		if s := ctx.Query("id"); s != "" {
			idValue, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, s, err).Error())
				return
			}
			id.Int64 = idValue
			id.Valid = true
		}
		var id sql.NullInt32
		if s := ctx.Query("id"); s != "" {
			idValue, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, s, err).Error())
				return
			}
			id.Int64 = int32(idValue)
			id.Valid = true
		}


		情况8
		var id sql.NullInt64
		id.Int64, err = ctx.GetInt64Query("id")
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, ctx.Param("id"), err).Error())
			return
		}
		id.Valid = true


		var id sql.NullString
		if s := ctx.GetQuery("id");  s != "" {
			id.Valid = true
			id.String = s
		}



		情况9
		// 这个情况不存在，
		var a *string = &ctx.Param("a")
		// 国为可以改成调用时用 & 
		//  a = &ctx.Param("a")
		//  func(&a)


		情况10
		var a *string
		if s := ctx.Query("a"); s != "" {
			a = &s
		}


		情况12
		var a *int
		if aValue, ok := ctx.GetIntParam("a"); !ok {
			ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", a, ctx.Param("a"), err).Error())
			return
		} else {
			a = &aValue
		}


		情况13
		var a *int
		if aValue, err := strconv.Atoi(ctx.Param("a")); err != nil {
			ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", a, ctx.Param("a"), err).Error())
			return
		} else {
			a = &aValue
		}


		情况14
		var a *int
		if s := ctx.Query("a"); s != "" {
			aValue, err := strconv.Atoi(s)
			if err != nil {
				ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
				return
			}
			a = &aValue
		}

		var a *int32
		if s := ctx.Query("a"); s != "" {
			aValue, err := strconv.Atoi(s)
			if err != nil {
				ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", a, s, err).Error())
				return
			}
			a = new(int32)
			*a = int32(aValue)
		}
		var a *int
		if aValue, ok := ctx.GetIntQuery("a"); ok {
			a = &aValue
		}



