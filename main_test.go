package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {

	handler := http.HandlerFunc(mainHandle)
	cityName := "tula"
	// Получаем общее количество кафе в городе
	cityCafes, exists := cafeList[cityName]
	if !exists {
		t.Fatalf("Город %s не найден в cafeList", cityName)
	}
	require.True(t, exists, "Город %s должен существовать в cafeList", cityName)
	totalCafesInCity := len(cityCafes)

	requests := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(100, totalCafesInCity)},
	}

	for _, v := range requests {
		t.Run(fmt.Sprintf("count=%d", v.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city="+cityName+"&count="+strconv.Itoa(v.count), nil)

			handler.ServeHTTP(response, req)

			// Проверяем, что запрос успешно обработан
			require.Equal(t, http.StatusOK, response.Code,
				"Для count=%d ожидался статус 200 OK, получен %d",
				v.count, response.Code)

			// Получаем тело ответа
			body := response.Body.String()
			body = strings.TrimSpace(body)

			// Обрабатываем пустой ответ
			if v.want == 0 {
				assert.Empty(t, body,
					"Для count=%d ожидался пустой ответ, получено: %s",
					v.count, body)
				return
			}

			cafes := strings.Split(body, ",")

			// Проверяем количество возвращаемых кафе
			assert.Equal(t, v.want, len(cafes),
				"Для count=%d ожидалось %d кафе, получено %d. Ответ: %s",
				v.count, v.want, len(cafes), body)

			// Дополнительная проверка: все элементы не пустые
			for i, cafe := range cafes {
				cafe = strings.TrimSpace(cafe)
				assert.NotEmpty(t, cafe,
					"Для count=%d кафе №%d пустое", v.count, i+1)
			}
		})
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, v := range requests {
		t.Run(fmt.Sprintf("search=%s", v.search), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city=moscow&search="+v.search, nil)

			handler.ServeHTTP(response, req)

			// Проверяем, что запрос успешно обработан
			require.Equal(t, http.StatusOK, response.Code,
				"Для search=%s ожидался статус 200 OK, получен %d",
				v.search, response.Code)

			// Получаем тело ответа
			body := response.Body.String()
			body = strings.TrimSpace(body)

			// Обрабатываем пустой ответ
			if v.wantCount == 0 {
				assert.Empty(t, body,
					"Для search=%s ожидался пустой ответ, получено: %s",
					v.search, body)
				return
			}

			cafes := strings.Split(body, ",")

			// Проверяем количество возвращаемых кафе
			assert.Equal(t, v.wantCount, len(cafes),
				"Для search=%s ожидалось %d кафе, получено %d.",
				v.search, v.wantCount, len(cafes), body)

			searchLower := strings.ToLower(v.search)

			for i, cafe := range cafes {
				cafe = strings.TrimSpace(cafe)
				cafeLower := strings.ToLower(cafe)

				// Проверяем, что кафе содержит искомую строку
				assert.True(t, strings.Contains(cafeLower, searchLower),
					"Для search='%s' кафе '%s' (№%d) не содержит искомую строку",
					v.search, cafe, i+1)
			}
		})
	}

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
