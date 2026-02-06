package digest

import (
	"net/http"
	"strconv"
	"time"

	cdigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/base"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var (
	HandlerPathStorageDesign      = `/storage/{contract:(?i)` + ctypes.REStringAddressString + `}`
	HandlerPathStorageData        = `/storage/{contract:(?i)` + ctypes.REStringAddressString + `}/datakey/{data_key:` + ctypes.ReSpecialCh + `}`
	HandlerPathStorageDataHistory = `/storage/{contract:(?i)` + ctypes.REStringAddressString + `}/datakey/{data_key:` + ctypes.ReSpecialCh + `}/history`
	HandlerPathStorageDataCount   = `/storage/{contract:(?i)` + ctypes.REStringAddressString + `}/datacount`
)

func SetHandlers(hd *cdigest.Handlers) {
	get := 1000
	_ = hd.SetHandler(HandlerPathStorageData, HandleStorageData, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathStorageDesign, HandleStorageDesign, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathStorageDataHistory, HandleStorageDataHistory, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathStorageDataCount, HandleStorageDataCount, true, get, get).
		Methods(http.MethodOptions, "GET")
}

func HandleStorageDesign(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleStorageDesignInGroup(hd, contract)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)

		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleStorageDesignInGroup(hd *cdigest.Handlers, contract string) ([]byte, error) {
	var de types.Design
	var st base.State

	de, st, err := StorageDesign(hd.Database(), contract)
	if err != nil {
		return nil, err
	}

	i, err := buildStorageDesign(hd, contract, de, st)
	if err != nil {
		return nil, err
	}
	return hd.Encoder().Marshal(i)
}

func buildStorageDesign(hd *cdigest.Handlers, contract string, de types.Design, st base.State) (cdigest.Hal, error) {
	h, err := hd.CombineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}

	var hal cdigest.Hal
	hal = cdigest.NewBaseHal(de, cdigest.NewHalLink(h, nil))

	h, err = hd.CombineURL(cdigest.HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", cdigest.NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.CombineURL(cdigest.HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", cdigest.NewHalLink(h, nil))
	}

	return hal, nil
}

func HandleStorageData(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cacheKey := cdigest.CacheKeyPath(r)
	if err := cdigest.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	key, err, status := cdigest.ParseRequest(w, r, "data_key")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleStorageDataInGroup(hd, contract, key)
	}); err != nil {
		cdigest.HTTP2HandleError(w, err)
	} else {
		cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)

		if !shared {
			cdigest.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleStorageDataInGroup(hd *cdigest.Handlers, contract, key string) ([]byte, error) {
	data, height, operation, timestamp, deleted, err := StorageData(hd.Database(), contract, key)
	if err != nil {
		return nil, err
	}

	i, err := buildStorageDataHal(hd, contract, *data, height, operation, timestamp, deleted)
	if err != nil {
		return nil, err
	}
	return hd.Encoder().Marshal(i)
}

func buildStorageDataHal(
	hd *cdigest.Handlers,
	contract string, data types.Data, height int64, operation, timestamp string, deleted bool) (cdigest.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathStorageData,
		"contract", contract, "data_key", data.DataKey())
	if err != nil {
		return nil, err
	}

	var hal cdigest.Hal
	hal = cdigest.NewBaseHal(
		struct {
			Data      types.Data `json:"data"`
			Height    int64      `json:"height"`
			Operation string     `json:"operation"`
			Timestamp string     `json:"timestamp"`
		}{Data: data, Height: height, Operation: operation, Timestamp: timestamp},
		cdigest.NewHalLink(h, nil),
	)

	h, err = hd.CombineURL(cdigest.HandlerPathBlockByHeight, "height", strconv.FormatInt(height, 10))
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", cdigest.NewHalLink(h, nil))

	h, err = hd.CombineURL(cdigest.HandlerPathOperation, "hash", operation)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("operation", cdigest.NewHalLink(h, nil))

	return hal, nil
}

func HandleStorageDataHistory(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	limit := cdigest.ParseLimitQuery(r.URL.Query().Get("limit"))
	offset := cdigest.ParseStringQuery(r.URL.Query().Get("offset"))
	reverse := cdigest.ParseBoolQuery(r.URL.Query().Get("reverse"))

	cacheKey := cdigest.CacheKey(
		r.URL.Path, cdigest.StringOffsetQuery(offset),
		cdigest.StringBoolQuery("reverse", reverse),
	)

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	key, err, status := cdigest.ParseRequest(w, r, "data_key")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		i, filled, err := handleStorageDataHistoryInGroup(hd, contract, key, offset, reverse, limit)

		return []interface{}{i, filled}, err
	})

	if err != nil {
		hd.Log().Err(err).Str("contract", contract).Msg("failed to get data history")
		cdigest.HTTP2HandleError(w, err)

		return
	}

	var b []byte
	var filled bool
	{
		l := v.([]interface{})
		b = l[0].([]byte)
		filled = l[1].(bool)
	}

	cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, b, http.StatusOK)

	if !shared {
		expire := hd.ExpireNotFilled()
		if len(offset) > 0 && filled {
			expire = time.Minute
		}

		cdigest.HTTP2WriteCache(w, cacheKey, expire)
	}
}

func handleStorageDataHistoryInGroup(
	hd *cdigest.Handlers,
	contract, key string,
	offset string,
	reverse bool,
	l int64,
) ([]byte, bool, error) {
	var limit int64
	if l < 0 {
		limit = hd.ItemsLimiter("service-credentials")
	} else {
		limit = l
	}

	var vas []cdigest.Hal
	if err := SotrageDataHistoryByDataKey(
		hd.Database(), contract, key, reverse, offset, limit,
		func(data *types.Data, height int64, operation, timestamp string, deleted bool) (bool, error) {
			hal, err := buildStorageDataHal(hd, contract, *data, height, operation, timestamp, deleted)
			if err != nil {
				return false, err
			}
			vas = append(vas, hal)

			return true, nil
		},
	); err != nil {
		return nil, false, mitumutil.ErrNotFound.WithMessage(err, "data history by contract %s, data key %s", contract, key)
	} else if len(vas) < 1 {
		return nil, false, mitumutil.ErrNotFound.Errorf("data history by contract %s, data key %s", contract, key)
	}

	i, err := buildStorageDataHistoryHal(hd, contract, key, vas, offset, reverse)
	if err != nil {
		return nil, false, err
	}

	b, err := hd.Encoder().Marshal(i)
	return b, int64(len(vas)) == limit, err
}

func buildStorageDataHistoryHal(
	hd *cdigest.Handlers,
	contract, key string,
	vas []cdigest.Hal,
	offset string,
	reverse bool,
) (cdigest.Hal, error) {
	baseSelf, err := hd.CombineURL(
		HandlerPathStorageDataHistory,
		"contract", contract,
		"data_key", key,
	)
	if err != nil {
		return nil, err
	}

	self := baseSelf
	if len(offset) > 0 {
		self = cdigest.AddQueryValue(baseSelf, cdigest.StringOffsetQuery(offset))
	}
	if reverse {
		self = cdigest.AddQueryValue(baseSelf, cdigest.StringBoolQuery("reverse", reverse))
	}

	var hal cdigest.Hal
	hal = cdigest.NewBaseHal(vas, cdigest.NewHalLink(self, nil))

	h, err := hd.CombineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("service", cdigest.NewHalLink(h, nil))

	var nextOffset string

	if len(vas) > 0 {
		va, ok := vas[len(vas)-1].Interface().(struct {
			Data      types.Data `json:"data"`
			Height    int64      `json:"height"`
			Operation string     `json:"operation"`
			Timestamp string     `json:"timestamp"`
		})
		if !ok {
			return nil, errors.Errorf("failed to build storage data history hal")
		}
		nextOffset = strconv.FormatInt(va.Height, 10)
	}

	if len(nextOffset) > 0 {
		next := baseSelf
		next = cdigest.AddQueryValue(next, cdigest.StringOffsetQuery(nextOffset))

		if reverse {
			next = cdigest.AddQueryValue(next, cdigest.StringBoolQuery("reverse", reverse))
		}

		hal = hal.AddLink("next", cdigest.NewHalLink(next, nil))
	}

	hal = hal.AddLink("reverse", cdigest.NewHalLink(cdigest.AddQueryValue(baseSelf, cdigest.StringBoolQuery("reverse", !reverse)), nil))

	return hal, nil
}

func HandleStorageDataCount(hd *cdigest.Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cachekey := cdigest.CacheKey(
		r.URL.Path,
	)

	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)

		return
	}

	deleted := cdigest.ParseBoolQuery(r.URL.Query().Get("deleted"))

	v, err, shared := hd.RG().Do(cachekey, func() (interface{}, error) {
		i, err := handleStorageDataCountInGroup(hd, contract, deleted)

		return i, err
	})

	if err != nil {
		hd.Log().Err(err).Str("contract", contract).Msg("failed to count data")
		cdigest.HTTP2HandleError(w, err)

		return
	}

	cdigest.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)

	if !shared {
		expire := hd.ExpireNotFilled()
		cdigest.HTTP2WriteCache(w, cachekey, expire)
	}
}

func handleStorageDataCountInGroup(
	hd *cdigest.Handlers,
	contract string, deleted bool,
) ([]byte, error) {
	count, err := DataCountByContract(
		hd.Database(), contract, deleted,
	)
	if err != nil {
		return nil, mitumutil.ErrNotFound.WithMessage(err, "data count by contract, %s", contract)
	}

	i, err := buildStorageDataCountHal(hd, contract, count)
	if err != nil {
		return nil, err
	}

	b, err := hd.Encoder().Marshal(i)
	return b, err
}

func buildStorageDataCountHal(
	hd *cdigest.Handlers,
	contract string,
	count int64,
) (cdigest.Hal, error) {
	baseSelf, err := hd.CombineURL(HandlerPathStorageDataCount, "contract", contract)
	if err != nil {
		return nil, err
	}

	self := baseSelf

	var m struct {
		Contract  string `json:"contract"`
		DataCount int64  `json:"data_count"`
	}

	m.Contract = contract
	m.DataCount = count

	var hal cdigest.Hal
	hal = cdigest.NewBaseHal(m, cdigest.NewHalLink(self, nil))

	h, err := hd.CombineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("collection", cdigest.NewHalLink(h, nil))

	return hal, nil
}
