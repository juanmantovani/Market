package app

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewSeedsDemoDataAndServesTheCategoryAPI(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("SEED_DEMO_DATA", "true")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	api := application.server.Handler

	listed := httptest.NewRecorder()
	api.ServeHTTP(listed, httptest.NewRequest(http.MethodGet, "/api/v1/categories?q=bebidas", nil))
	var page struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(listed.Body).Decode(&page); err != nil || listed.Code != http.StatusOK {
		t.Fatalf("GET categories = %d, %v", listed.Code, err)
	}
	if page.Total != 1 || page.Items[0].Name != "Bebidas" {
		t.Fatalf("search for bebidas = %+v", page)
	}

	duplicate := httptest.NewRecorder()
	api.ServeHTTP(duplicate, httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"BEBIDAS"}`)))
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("POST duplicate = %d, want 409", duplicate.Code)
	}

	bebidasItems := httptest.NewRecorder()
	api.ServeHTTP(bebidasItems, httptest.NewRequest(http.MethodGet, "/api/v1/items?categoryId="+page.Items[0].ID, nil))
	if bebidasItems.Code != http.StatusOK || !strings.Contains(bebidasItems.Body.String(), `"total":2`) {
		t.Fatalf("GET items of Bebidas = %d %s", bebidasItems.Code, bebidasItems.Body.String())
	}

	inUse := httptest.NewRecorder()
	api.ServeHTTP(inUse, httptest.NewRequest(http.MethodDelete, "/api/v1/categories/"+page.Items[0].ID, nil))
	if inUse.Code != http.StatusConflict || !strings.Contains(inUse.Body.String(), "category_in_use") {
		t.Fatalf("DELETE category with items = %d %s, want 409 category_in_use", inUse.Code, inUse.Body.String())
	}

	created := httptest.NewRecorder()
	api.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"Ferretería"}`)))
	var ferreteria struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(created.Body).Decode(&ferreteria); err != nil || created.Code != http.StatusCreated {
		t.Fatalf("POST category = %d, %v", created.Code, err)
	}

	deleted := httptest.NewRecorder()
	api.ServeHTTP(deleted, httptest.NewRequest(http.MethodDelete, "/api/v1/categories/"+ferreteria.ID, nil))
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("DELETE empty category = %d %s", deleted.Code, deleted.Body.String())
	}

	gone := httptest.NewRecorder()
	api.ServeHTTP(gone, httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+ferreteria.ID, nil))
	if gone.Code != http.StatusNotFound {
		t.Fatalf("GET after DELETE = %d, want 404", gone.Code)
	}
}

func TestStockFlowOverTheDemoData(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("SEED_DEMO_DATA", "true")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	call := func(method, target, body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		application.server.Handler.ServeHTTP(recorder, httptest.NewRequest(method, target, strings.NewReader(body)))
		return recorder
	}
	type level struct {
		ItemID   string `json:"itemId"`
		Code     string `json:"code"`
		Quantity int64  `json:"quantity"`
	}
	lowStock := func() []level {
		recorder := call(http.MethodGet, "/api/v1/stock?lowStock=true", "")
		var page struct {
			Items []level `json:"items"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&page); err != nil || recorder.Code != http.StatusOK {
			t.Fatalf("GET low stock = %d, %v", recorder.Code, err)
		}
		return page.Items
	}

	low := lowStock()
	if len(low) != 3 || low[2].Code != "ALM-001" || low[2].Quantity != 4_000 {
		t.Fatalf("low stock = %+v, want leche, queso and yerba", low)
	}
	yerba := low[2].ItemID

	entry := call(http.MethodPost, "/api/v1/stock/entries", `{"lines":[{"itemId":"`+yerba+`","quantity":12000}]}`)
	if entry.Code != http.StatusCreated || !strings.Contains(entry.Body.String(), `"balanceAfter":16000`) {
		t.Fatalf("POST entry = %d %s", entry.Code, entry.Body.String())
	}
	if after := lowStock(); len(after) != 2 {
		t.Fatalf("low stock after the entry = %+v, want yerba gone", after)
	}

	tooMuch := call(http.MethodPost, "/api/v1/stock/adjustments", `{"itemId":"`+yerba+`","quantity":-17000,"reason":"Conteo"}`)
	if tooMuch.Code != http.StatusConflict || !strings.Contains(tooMuch.Body.String(), "insufficient_stock") {
		t.Fatalf("POST adjustment below zero = %d %s", tooMuch.Code, tooMuch.Body.String())
	}
	adjusted := call(http.MethodPost, "/api/v1/stock/adjustments", `{"itemId":"`+yerba+`","quantity":-1000,"reason":"Rotura"}`)
	if adjusted.Code != http.StatusCreated || !strings.Contains(adjusted.Body.String(), `"balanceAfter":15000`) {
		t.Fatalf("POST adjustment = %d %s", adjusted.Code, adjusted.Body.String())
	}

	history := call(http.MethodGet, "/api/v1/stock/movements?itemId="+yerba, "")
	var movements struct {
		Items []struct {
			Type string `json:"type"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(history.Body).Decode(&movements); err != nil || movements.Total != 3 {
		t.Fatalf("GET movements = %+v, %v", movements, err)
	}
	if movements.Items[0].Type != "adjustment" || movements.Items[1].Type != "entry" || movements.Items[2].Type != "initial" {
		t.Fatalf("movements = %+v, want newest first", movements.Items)
	}

	inUse := call(http.MethodDelete, "/api/v1/items/"+yerba, "")
	if inUse.Code != http.StatusConflict || !strings.Contains(inUse.Body.String(), "item_in_use") {
		t.Fatalf("DELETE item with movements = %d %s", inUse.Code, inUse.Body.String())
	}
}

func TestNewStartsEmptyWithoutDemoData(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("SEED_DEMO_DATA", "false")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	application.server.Handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil))
	if !strings.Contains(recorder.Body.String(), `"total":0`) {
		t.Fatalf("GET categories without demo data = %s", recorder.Body.String())
	}
}

func freePort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("finding a free port: %v", err)
	}
	defer listener.Close()

	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
}

func TestNewServesLivenessThroughTheMiddlewareChain(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://localhost:5173")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	application.server.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatal("CORS middleware is not wired")
	}
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	t.Setenv("STORAGE_DRIVER", "sqlite")

	if _, err := New(context.Background()); err == nil {
		t.Fatal("New() error = nil, want a configuration error")
	}
}

func TestRunShutsDownWhenTheContextIsCancelled(t *testing.T) {
	t.Setenv("PORT", freePort(t))
	t.Setenv("STORAGE_DRIVER", "memory")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return after cancellation")
	}
}

type flow struct {
	t       *testing.T
	handler http.Handler
}

func newFlow(t *testing.T, allowNegative string) flow {
	t.Helper()
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("SEED_DEMO_DATA", "false")
	t.Setenv("STOCK_ALLOW_NEGATIVE", allowNegative)

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return flow{t: t, handler: application.server.Handler}
}

func (flow flow) call(method, target, key, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	recorder := httptest.NewRecorder()
	flow.handler.ServeHTTP(recorder, request)
	return recorder
}

// create posts body and returns the id of what it created.
func (flow flow) create(target, body string) string {
	flow.t.Helper()
	recorder := flow.call(http.MethodPost, target, "", body)
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&created); err != nil || recorder.Code != http.StatusCreated {
		flow.t.Fatalf("POST %s = %d, %v", target, recorder.Code, err)
	}
	return created.ID
}

func (flow flow) stockOf(itemID string) int64 {
	flow.t.Helper()
	recorder := flow.call(http.MethodGet, "/api/v1/stock?limit=100", "", "")
	var page struct {
		Items []struct {
			ItemID   string `json:"itemId"`
			Quantity int64  `json:"quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&page); err != nil {
		flow.t.Fatalf("GET stock: %v", err)
	}
	for _, level := range page.Items {
		if level.ItemID == itemID {
			return level.Quantity
		}
	}
	flow.t.Fatalf("item %s has no stock row", itemID)
	return 0
}

func (flow flow) expect(recorder *httptest.ResponseRecorder, status int, fragments ...string) {
	flow.t.Helper()
	if recorder.Code != status {
		flow.t.Fatalf("status = %d, want %d (body %s)", recorder.Code, status, recorder.Body.String())
	}
	for _, fragment := range fragments {
		if !strings.Contains(recorder.Body.String(), fragment) {
			flow.t.Fatalf("body %s does not contain %s", recorder.Body.String(), fragment)
		}
	}
}

func TestSaleFlow(t *testing.T) {
	flow := newFlow(t, "")
	categoryID := flow.create("/api/v1/categories", `{"name":"Bebidas"}`)
	itemID := flow.create("/api/v1/items", `{"code":"BEB-1","name":"Gaseosa cola","categoryId":"`+categoryID+`","unit":"unit","salePrice":150000,"taxRate":2100,"minStock":0}`)
	flow.expect(flow.call(http.MethodPost, "/api/v1/stock/entries", "", `{"lines":[{"itemId":"`+itemID+`","quantity":5000}]}`), http.StatusCreated)
	threeColas := `{"lines":[{"itemId":"` + itemID + `","quantity":3000,"discountPercent":0}],"payments":[{"method":"cash","amount":500000}],"expectedTotal":450000}`

	flow.expect(flow.call(http.MethodPost, "/api/v1/sales", "", threeColas), http.StatusBadRequest, "idempotency_key_required")

	sold := flow.call(http.MethodPost, "/api/v1/sales", "sale-1", threeColas)
	flow.expect(sold, http.StatusCreated, `"number":1`, `"total":450000`, `"change":50000`, `"name":"Consumidor Final"`, `"seller":null`, `"status":"completed"`)
	if stock := flow.stockOf(itemID); stock != 2_000 {
		t.Fatalf("stock after the sale = %d, want 2000", stock)
	}

	retried := flow.call(http.MethodPost, "/api/v1/sales", "sale-1", threeColas)
	flow.expect(retried, http.StatusCreated)
	if retried.Body.String() != sold.Body.String() {
		t.Fatalf("retry = %s, want the first sale %s", retried.Body.String(), sold.Body.String())
	}
	if stock := flow.stockOf(itemID); stock != 2_000 {
		t.Fatalf("stock after a retry = %d, want 2000", stock)
	}
	flow.expect(flow.call(http.MethodPost, "/api/v1/sales", "sale-1", strings.Replace(threeColas, "500000", "450000", 1)), http.StatusConflict, "idempotency_key_conflict")

	flow.expect(flow.call(http.MethodPost, "/api/v1/sales", "sale-2", threeColas), http.StatusConflict, "insufficient_stock", "Gaseosa cola")
	flow.expect(flow.call(http.MethodPost, "/api/v1/sales", "sale-3", strings.Replace(threeColas, `"expectedTotal":450000`, `"expectedTotal":400000`, 1)), http.StatusConflict, "sale_total_mismatch")
	if stock := flow.stockOf(itemID); stock != 2_000 {
		t.Fatalf("stock after rejected sales = %d, want 2000", stock)
	}

	customerID := flow.create("/api/v1/customers", `{"name":"Ana Pérez","documentType":"dni","documentNumber":"28456789","taxCondition":"final_consumer"}`)
	sellerID := flow.create("/api/v1/sellers", `{"name":"Juan Pérez","code":"JP"}`)
	twoWithDiscount := `{"customerId":"` + customerID + `","sellerId":"` + sellerID + `","lines":[{"itemId":"` + itemID + `","quantity":2000,"discountPercent":1000}],` +
		`"payments":[{"method":"debit_card","amount":270000,"reference":"Cupón 1"}],"expectedTotal":270000}`
	flow.expect(flow.call(http.MethodPost, "/api/v1/sales", "sale-4", twoWithDiscount), http.StatusCreated,
		`"number":2`, `"discountTotal":30000`, `"name":"Ana Pérez"`, `"code":"JP"`, `"change":0`)
	if stock := flow.stockOf(itemID); stock != 0 {
		t.Fatalf("stock after selling everything = %d, want 0", stock)
	}

	flow.expect(flow.call(http.MethodDelete, "/api/v1/customers/"+customerID, "", ""), http.StatusConflict, "customer_in_use")
	flow.expect(flow.call(http.MethodDelete, "/api/v1/sellers/"+sellerID, "", ""), http.StatusConflict, "seller_in_use")
	flow.expect(flow.call(http.MethodDelete, "/api/v1/customers/00000000-0000-7000-8000-000000000001", "", ""), http.StatusConflict, "customer_is_system")
	flow.expect(flow.call(http.MethodPut, "/api/v1/customers/00000000-0000-7000-8000-000000000001", "", `{"name":"Otro","documentType":"none","taxCondition":"final_consumer"}`), http.StatusConflict, "customer_is_system")
	flow.expect(flow.call(http.MethodGet, "/api/v1/customers?q=consumidor", "", ""), http.StatusOK, `"system":true`)

	flow.expect(flow.call(http.MethodGet, "/api/v1/stock/movements?itemId="+itemID+"&type=sale", "", ""), http.StatusOK, `"total":2`, `"reference":"Venta 2"`, `"reference":"Venta 1"`)

	var first struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(strings.NewReader(sold.Body.String())).Decode(&first); err != nil {
		t.Fatalf("decoding the first sale: %v", err)
	}
	flow.expect(flow.call(http.MethodGet, "/api/v1/sales/"+first.ID, "", ""), http.StatusOK, `"number":1`, `"voidedAt":null`)
	flow.expect(flow.call(http.MethodGet, "/api/v1/sales?customerId="+customerID, "", ""), http.StatusOK, `"total":1`, `"number":2`)

	flow.expect(flow.call(http.MethodPost, "/api/v1/sales/"+first.ID+"/void", "", `{"reason":" "}`), http.StatusBadRequest, "sale_void_reason_required")
	flow.expect(flow.call(http.MethodPost, "/api/v1/sales/"+first.ID+"/void", "", `{"reason":"Error de carga"}`), http.StatusOK,
		`"status":"voided"`, `"voidReason":"Error de carga"`, `"voidedBy":"dev"`)
	if stock := flow.stockOf(itemID); stock != 3_000 {
		t.Fatalf("stock after voiding the first sale = %d, want 3000", stock)
	}
	flow.expect(flow.call(http.MethodPost, "/api/v1/sales/"+first.ID+"/void", "", `{"reason":"Otra vez"}`), http.StatusConflict, "sale_already_voided")
	if stock := flow.stockOf(itemID); stock != 3_000 {
		t.Fatalf("stock after voiding twice = %d, want 3000", stock)
	}

	flow.expect(flow.call(http.MethodGet, "/api/v1/sales?status=voided", "", ""), http.StatusOK, `"total":1`, `"number":1`)
	flow.expect(flow.call(http.MethodGet, "/api/v1/sales?status=completed", "", ""), http.StatusOK, `"total":1`, `"number":2`)
	flow.expect(flow.call(http.MethodGet, "/api/v1/sales?from=2000-01-01T00:00:00Z&to=2000-01-02T00:00:00Z", "", ""), http.StatusOK, `"total":0`)
	flow.expect(flow.call(http.MethodGet, "/api/v1/stock/movements?itemId="+itemID+"&type=sale_void", "", ""), http.StatusOK, `"total":1`, `"quantity":3000`, `"reference":"Venta 1"`)
}

func TestSaleFlowWithNegativeStockAllowed(t *testing.T) {
	flow := newFlow(t, "true")
	categoryID := flow.create("/api/v1/categories", `{"name":"Fiambrería"}`)
	itemID := flow.create("/api/v1/items", `{"code":"FIA-1","name":"Queso cremoso","categoryId":"`+categoryID+`","unit":"kg","salePrice":980000,"taxRate":2100,"minStock":0}`)

	flow.expect(flow.call(http.MethodPost, "/api/v1/sales", "sale-1",
		`{"lines":[{"itemId":"`+itemID+`","quantity":250}],"payments":[{"method":"qr","amount":245000}],"expectedTotal":245000}`), http.StatusCreated)

	if stock := flow.stockOf(itemID); stock != -250 {
		t.Fatalf("stock = %d, want -250", stock)
	}
}
