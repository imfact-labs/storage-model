package api

import (
	"net/http"
	"strconv"
	"time"

	apic "github.com/imfact-labs/currency-model/api"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	mitumutil "github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/storage-model/digest"
	"github.com/imfact-labs/storage-model/types"
	"github.com/pkg/errors"
)

var (
	HandlerPathStorageDesign      = `/storage/{contract:(?i)` + ctypes.REStringAddressString + `}`
	HandlerPathStorageData        = `/storage/{contract:(?i)` + ctypes.REStringAddressString + `}/datakey/{data_key:` + ctypes.ReSpecialCh + `}`
	HandlerPathStorageDataHistory = `/storage/{contract:(?i)` + ctypes.REStringAddressString + `}/datakey/{data_key:` + ctypes.ReSpecialCh + `}/history`
	HandlerPathStorageDataCount   = `/storage/{contract:(?i)` + ctypes.REStringAddressString + `}/datacount`
)

func SetHandlers(hd *apic.Handlers) {
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

func HandleStorageDesign(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cacheKey := apic.CacheKeyPath(r)
	if err := apic.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleStorageDesignInGroup(hd, contract)
	}); err != nil {
		apic.HTTP2HandleError(w, err)
	} else {
		apic.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)

		if !shared {
			apic.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleStorageDesignInGroup(hd *apic.Handlers, contract string) ([]byte, error) {
	var de types.Design
	var st base.State

	de, st, err := digest.StorageDesign(hd.Database(), contract)
	if err != nil {
		return nil, err
	}

	i, err := buildStorageDesign(hd, contract, de, st)
	if err != nil {
		return nil, err
	}
	return hd.Encoder().Marshal(i)
}

func buildStorageDesign(hd *apic.Handlers, contract string, de types.Design, st base.State) (apic.Hal, error) {
	h, err := hd.CombineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}

	var hal apic.Hal
	hal = apic.NewBaseHal(de, apic.NewHalLink(h, nil))

	h, err = hd.CombineURL(apic.HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", apic.NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.CombineURL(apic.HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", apic.NewHalLink(h, nil))
	}

	return hal, nil
}

func HandleStorageData(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cacheKey := apic.CacheKeyPath(r)
	if err := apic.LoadFromCache(hd.Cache(), cacheKey, w); err == nil {
		return
	}

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	key, err, status := apic.ParseRequest(w, r, "data_key")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		return handleStorageDataInGroup(hd, contract, key)
	}); err != nil {
		apic.HTTP2HandleError(w, err)
	} else {
		apic.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)

		if !shared {
			apic.HTTP2WriteCache(w, cacheKey, hd.ExpireShortLived())
		}
	}
}

func handleStorageDataInGroup(hd *apic.Handlers, contract, key string) ([]byte, error) {
	data, height, operation, timestamp, deleted, err := digest.StorageData(hd.Database(), contract, key)
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
	hd *apic.Handlers,
	contract string, data types.Data, height int64, operation, timestamp string, deleted bool) (apic.Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathStorageData,
		"contract", contract, "data_key", data.DataKey())
	if err != nil {
		return nil, err
	}

	var hal apic.Hal
	hal = apic.NewBaseHal(
		struct {
			Data      types.Data `json:"data"`
			Height    int64      `json:"height"`
			Operation string     `json:"operation"`
			Timestamp string     `json:"timestamp"`
		}{Data: data, Height: height, Operation: operation, Timestamp: timestamp},
		apic.NewHalLink(h, nil),
	)

	h, err = hd.CombineURL(apic.HandlerPathBlockByHeight, "height", strconv.FormatInt(height, 10))
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", apic.NewHalLink(h, nil))

	h, err = hd.CombineURL(apic.HandlerPathOperation, "hash", operation)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("operation", apic.NewHalLink(h, nil))

	return hal, nil
}

func HandleStorageDataHistory(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	limit := apic.ParseLimitQuery(r.URL.Query().Get("limit"))
	offset := apic.ParseStringQuery(r.URL.Query().Get("offset"))
	reverse := apic.ParseBoolQuery(r.URL.Query().Get("reverse"))

	cacheKey := apic.CacheKey(
		r.URL.Path, apic.StringOffsetQuery(offset),
		apic.StringBoolQuery("reverse", reverse),
	)

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	key, err, status := apic.ParseRequest(w, r, "data_key")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)
		return
	}

	v, err, shared := hd.RG().Do(cacheKey, func() (interface{}, error) {
		i, filled, err := handleStorageDataHistoryInGroup(hd, contract, key, offset, reverse, limit)

		return []interface{}{i, filled}, err
	})

	if err != nil {
		hd.Log().Err(err).Str("contract", contract).Msg("failed to get data history")
		apic.HTTP2HandleError(w, err)

		return
	}

	var b []byte
	var filled bool
	{
		l := v.([]interface{})
		b = l[0].([]byte)
		filled = l[1].(bool)
	}

	apic.HTTP2WriteHalBytes(hd.Encoder(), w, b, http.StatusOK)

	if !shared {
		expire := hd.ExpireNotFilled()
		if len(offset) > 0 && filled {
			expire = time.Minute
		}

		apic.HTTP2WriteCache(w, cacheKey, expire)
	}
}

func handleStorageDataHistoryInGroup(
	hd *apic.Handlers,
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

	var vas []apic.Hal
	if err := digest.SotrageDataHistoryByDataKey(
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
	hd *apic.Handlers,
	contract, key string,
	vas []apic.Hal,
	offset string,
	reverse bool,
) (apic.Hal, error) {
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
		self = apic.AddQueryValue(baseSelf, apic.StringOffsetQuery(offset))
	}
	if reverse {
		self = apic.AddQueryValue(baseSelf, apic.StringBoolQuery("reverse", reverse))
	}

	var hal apic.Hal
	hal = apic.NewBaseHal(vas, apic.NewHalLink(self, nil))

	h, err := hd.CombineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("service", apic.NewHalLink(h, nil))

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
		next = apic.AddQueryValue(next, apic.StringOffsetQuery(nextOffset))

		if reverse {
			next = apic.AddQueryValue(next, apic.StringBoolQuery("reverse", reverse))
		}

		hal = hal.AddLink("next", apic.NewHalLink(next, nil))
	}

	hal = hal.AddLink("reverse", apic.NewHalLink(apic.AddQueryValue(baseSelf, apic.StringBoolQuery("reverse", !reverse)), nil))

	return hal, nil
}

func HandleStorageDataCount(hd *apic.Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cachekey := apic.CacheKey(
		r.URL.Path,
	)

	contract, err, status := apic.ParseRequest(w, r, "contract")
	if err != nil {
		apic.HTTP2ProblemWithError(w, err, status)

		return
	}

	deleted := apic.ParseBoolQuery(r.URL.Query().Get("deleted"))

	v, err, shared := hd.RG().Do(cachekey, func() (interface{}, error) {
		i, err := handleStorageDataCountInGroup(hd, contract, deleted)

		return i, err
	})

	if err != nil {
		hd.Log().Err(err).Str("contract", contract).Msg("failed to count data")
		apic.HTTP2HandleError(w, err)

		return
	}

	apic.HTTP2WriteHalBytes(hd.Encoder(), w, v.([]byte), http.StatusOK)

	if !shared {
		expire := hd.ExpireNotFilled()
		apic.HTTP2WriteCache(w, cachekey, expire)
	}
}

func handleStorageDataCountInGroup(
	hd *apic.Handlers,
	contract string, deleted bool,
) ([]byte, error) {
	count, err := digest.DataCountByContract(
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
	hd *apic.Handlers,
	contract string,
	count int64,
) (apic.Hal, error) {
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

	var hal apic.Hal
	hal = apic.NewBaseHal(m, apic.NewHalLink(self, nil))

	h, err := hd.CombineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("collection", apic.NewHalLink(h, nil))

	return hal, nil
}
