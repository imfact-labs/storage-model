package digest

import (
	currencydigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/base"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"time"
)

func (hd *Handlers) handleStorageDesign(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cacheKey := currencydigest.CacheKeyPath(r)
	if err := currencydigest.LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handleStorageDesignInGroup(contract)
	}); err != nil {
		currencydigest.HTTP2HandleError(w, err)
	} else {
		currencydigest.HTTP2WriteHalBytes(hd.encoder, w, v.([]byte), http.StatusOK)

		if !shared {
			currencydigest.HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handleStorageDesignInGroup(contract string) ([]byte, error) {
	var de types.Design
	var st base.State

	de, st, err := StorageDesign(hd.database, contract)
	if err != nil {
		return nil, err
	}

	i, err := hd.buildStorageDesign(contract, de, st)
	if err != nil {
		return nil, err
	}
	return hd.encoder.Marshal(i)
}

func (hd *Handlers) buildStorageDesign(contract string, de types.Design, st base.State) (currencydigest.Hal, error) {
	h, err := hd.combineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}

	var hal currencydigest.Hal
	hal = currencydigest.NewBaseHal(de, currencydigest.NewHalLink(h, nil))

	h, err = hd.combineURL(currencydigest.HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", currencydigest.NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.combineURL(currencydigest.HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", currencydigest.NewHalLink(h, nil))
	}

	return hal, nil
}

func (hd *Handlers) handleStorageData(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cacheKey := currencydigest.CacheKeyPath(r)
	if err := currencydigest.LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	key, err, status := currencydigest.ParseRequest(w, r, "data_key")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handleStorageDataInGroup(contract, key)
	}); err != nil {
		currencydigest.HTTP2HandleError(w, err)
	} else {
		currencydigest.HTTP2WriteHalBytes(hd.encoder, w, v.([]byte), http.StatusOK)

		if !shared {
			currencydigest.HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handleStorageDataInGroup(contract, key string) ([]byte, error) {
	data, height, operation, timestamp, deleted, err := StorageData(hd.database, contract, key)
	if err != nil {
		return nil, err
	}

	i, err := hd.buildStorageDataHal(contract, *data, height, operation, timestamp, deleted)
	if err != nil {
		return nil, err
	}
	return hd.encoder.Marshal(i)
}

func (hd *Handlers) buildStorageDataHal(
	contract string, data types.Data, height int64, operation, timestamp string, deleted bool) (currencydigest.Hal, error) {
	h, err := hd.combineURL(
		HandlerPathStorageData,
		"contract", contract, "data_key", data.DataKey())
	if err != nil {
		return nil, err
	}

	var hal currencydigest.Hal
	hal = currencydigest.NewBaseHal(
		struct {
			Data      types.Data `json:"data"`
			Height    int64      `json:"height"`
			Operation string     `json:"operation"`
			Timestamp string     `json:"timestamp"`
		}{Data: data, Height: height, Operation: operation, Timestamp: timestamp},
		currencydigest.NewHalLink(h, nil),
	)

	h, err = hd.combineURL(currencydigest.HandlerPathBlockByHeight, "height", strconv.FormatInt(height, 10))
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", currencydigest.NewHalLink(h, nil))

	h, err = hd.combineURL(currencydigest.HandlerPathOperation, "hash", operation)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("operation", currencydigest.NewHalLink(h, nil))

	return hal, nil
}

func (hd *Handlers) handleStorageDataHistory(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	limit := currencydigest.ParseLimitQuery(r.URL.Query().Get("limit"))
	offset := currencydigest.ParseStringQuery(r.URL.Query().Get("offset"))
	reverse := currencydigest.ParseBoolQuery(r.URL.Query().Get("reverse"))

	cacheKey := currencydigest.CacheKey(
		r.URL.Path, currencydigest.StringOffsetQuery(offset),
		currencydigest.StringBoolQuery("reverse", reverse),
	)

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	key, err, status := currencydigest.ParseRequest(w, r, "data_key")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		i, filled, err := hd.handleStorageDataHistoryInGroup(contract, key, offset, reverse, limit)

		return []interface{}{i, filled}, err
	})

	if err != nil {
		hd.Log().Err(err).Str("contract", contract).Msg("failed to get data history")
		currencydigest.HTTP2HandleError(w, err)

		return
	}

	var b []byte
	var filled bool
	{
		l := v.([]interface{})
		b = l[0].([]byte)
		filled = l[1].(bool)
	}

	currencydigest.HTTP2WriteHalBytes(hd.encoder, w, b, http.StatusOK)

	if !shared {
		expire := hd.expireNotFilled
		if len(offset) > 0 && filled {
			expire = time.Minute
		}

		currencydigest.HTTP2WriteCache(w, cacheKey, expire)
	}
}

func (hd *Handlers) handleStorageDataHistoryInGroup(
	contract, key string,
	offset string,
	reverse bool,
	l int64,
) ([]byte, bool, error) {
	var limit int64
	if l < 0 {
		limit = hd.itemsLimiter("service-credentials")
	} else {
		limit = l
	}

	var vas []currencydigest.Hal
	if err := SotrageDataHistoryByDataKey(
		hd.database, contract, key, reverse, offset, limit,
		func(data *types.Data, height int64, operation, timestamp string, deleted bool) (bool, error) {
			hal, err := hd.buildStorageDataHal(contract, *data, height, operation, timestamp, deleted)
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

	i, err := hd.buildStorageDataHistoryHal(contract, key, vas, offset, reverse)
	if err != nil {
		return nil, false, err
	}

	b, err := hd.encoder.Marshal(i)
	return b, int64(len(vas)) == limit, err
}

func (hd *Handlers) buildStorageDataHistoryHal(
	contract, key string,
	vas []currencydigest.Hal,
	offset string,
	reverse bool,
) (currencydigest.Hal, error) {
	baseSelf, err := hd.combineURL(
		HandlerPathStorageDataHistory,
		"contract", contract,
		"data_key", key,
	)
	if err != nil {
		return nil, err
	}

	self := baseSelf
	if len(offset) > 0 {
		self = currencydigest.AddQueryValue(baseSelf, currencydigest.StringOffsetQuery(offset))
	}
	if reverse {
		self = currencydigest.AddQueryValue(baseSelf, currencydigest.StringBoolQuery("reverse", reverse))
	}

	var hal currencydigest.Hal
	hal = currencydigest.NewBaseHal(vas, currencydigest.NewHalLink(self, nil))

	h, err := hd.combineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("service", currencydigest.NewHalLink(h, nil))

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
		next = currencydigest.AddQueryValue(next, currencydigest.StringOffsetQuery(nextOffset))

		if reverse {
			next = currencydigest.AddQueryValue(next, currencydigest.StringBoolQuery("reverse", reverse))
		}

		hal = hal.AddLink("next", currencydigest.NewHalLink(next, nil))
	}

	hal = hal.AddLink("reverse", currencydigest.NewHalLink(currencydigest.AddQueryValue(baseSelf, currencydigest.StringBoolQuery("reverse", !reverse)), nil))

	return hal, nil
}

func (hd *Handlers) handleStorageDataCount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cachekey := currencydigest.CacheKey(
		r.URL.Path,
	)

	contract, err, status := currencydigest.ParseRequest(w, r, "contract")
	if err != nil {
		currencydigest.HTTP2ProblemWithError(w, err, status)

		return
	}

	deleted := currencydigest.ParseBoolQuery(r.URL.Query().Get("deleted"))

	v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		i, err := hd.handleStorageDataCountInGroup(contract, deleted)

		return i, err
	})

	if err != nil {
		hd.Log().Err(err).Str("contract", contract).Msg("failed to count data")
		currencydigest.HTTP2HandleError(w, err)

		return
	}

	currencydigest.HTTP2WriteHalBytes(hd.encoder, w, v.([]byte), http.StatusOK)

	if !shared {
		expire := hd.expireNotFilled
		currencydigest.HTTP2WriteCache(w, cachekey, expire)
	}
}

func (hd *Handlers) handleStorageDataCountInGroup(
	contract string, deleted bool,
) ([]byte, error) {
	count, err := DataCountByContract(
		hd.database, contract, deleted,
	)
	if err != nil {
		return nil, mitumutil.ErrNotFound.WithMessage(err, "data count by contract, %s", contract)
	}

	i, err := hd.buildStorageDataCountHal(contract, count)
	if err != nil {
		return nil, err
	}

	b, err := hd.encoder.Marshal(i)
	return b, err
}

func (hd *Handlers) buildStorageDataCountHal(
	contract string,
	count int64,
) (currencydigest.Hal, error) {
	baseSelf, err := hd.combineURL(HandlerPathStorageDataCount, "contract", contract)
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

	var hal currencydigest.Hal
	hal = currencydigest.NewBaseHal(m, currencydigest.NewHalLink(self, nil))

	h, err := hd.combineURL(HandlerPathStorageDesign, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("collection", currencydigest.NewHalLink(h, nil))

	return hal, nil
}
