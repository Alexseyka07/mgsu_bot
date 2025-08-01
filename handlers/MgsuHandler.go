package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StudentInfo представляет информацию о позиции студента в конкурсном списке
type StudentInfo struct {
	BudgetPlaces    int    // Количество бюджетных мест
	Position        string // Позиция в формате "69/107"
	MinPassingScore int    // Минимальный проходной балл
	CreationDate    string // Дата создания списка
	CreationTime    string // Время создания списка
	Direction       string // Направление обучения
}

// StudentEntry представляет запись о студенте в таблице
type StudentEntry struct {
	Number                string
	UniqueCode            string
	TotalScore            string
	SubjectScore          string
	Math                  string
	IT                    string
	Russian               string
	GeneralAchievements   string
	AdmissionConsent      string
	Priority              string
	MainHighPriority      string
	IsMainHighPriority    string
	HighPassingPriority   string
	IsHighPassingPriority string
	PPR9                  string
	PPR10                 string
	BVIBasis              string
}

type MgsuHandler struct {
	botHandler           BotHandler
	lastCreationDateTime string
	subscribedUsers      map[int64]int // chatID -> uniqueCode
	mutex                sync.RWMutex
	monitoringActive     bool
}

func NewMgsuHandler(botHandler *BotHandler) MgsuHandler {
	return MgsuHandler{
		botHandler:       *botHandler,
		subscribedUsers:  make(map[int64]int),
		mutex:            sync.RWMutex{},
		monitoringActive: false,
	}
}

func (h *MgsuHandler) MgsuHandler(update *tgbotapi.Update) bool {
	if update.Message != nil {
		h.handleCommand(update.Message)
		return true
	}
	return false
}

func (h *MgsuHandler) handleCommand(message *tgbotapi.Message) {
	switch message.Text {
	case "Получить":
		h.handleGetCommand(message)
	case "Подписаться":
		h.handleSubscribeCommand(message)
	case "Отписаться":
		h.handleUnsubscribeCommand(message)
	}
}

func (h *MgsuHandler) handleGetCommand(message *tgbotapi.Message) {
	// Пример использования парсера
	// В реальной реализации uniqueCode должен приходить от пользователя
	uniqueCode := 3838475 // Пример кода из таблицы

	studentInfo, err := h.ParseStudentPosition(uniqueCode)
	if err != nil {
		msg := fmt.Sprintf("Ошибка при получении информации: %v", err)
		h.botHandler.SendTextMessage(message.Chat.ID, msg)
		return
	}

	msg := fmt.Sprintf(
		"Информация о студенте с кодом %d:\n"+
			"🎯 Позиция: %s\n"+
			"📚 Количество бюджетных мест: %d\n"+
			"📊 Минимальный проходной балл: %d\n"+
			"📅 Дата создания: %s\n"+
			"⏰ Время создания: %s\n"+
			"🎓 Направление: %s",
		uniqueCode,
		studentInfo.Position,
		studentInfo.BudgetPlaces,
		studentInfo.MinPassingScore,
		h.formatDate(studentInfo.CreationDate),
		studentInfo.CreationTime,
		studentInfo.Direction,
	)

	// Проверяем, подписан ли пользователь, и показываем соответствующие кнопки
	var buttons []string
	if h.IsSubscribed(message.Chat.ID) {
		buttons = []string{"Получить", "Отписаться"}
	} else {
		buttons = []string{"Получить", "Подписаться"}
	}

	commands := h.botHandler.SetKeyboardButtons(buttons, 2)
	h.botHandler.SendTextMessageWithKeyboardMarkup(message.Chat.ID, msg, commands)
}

// handleSubscribeCommand обрабатывает команду подписки на уведомления
func (h *MgsuHandler) handleSubscribeCommand(message *tgbotapi.Message) {
	// В реальной реализации uniqueCode должен запрашиваться у пользователя
	// Пока используем тестовый код
	uniqueCode := 3838475

	// Проверяем, не подписан ли пользователь уже
	if h.IsSubscribed(message.Chat.ID) {
		msg := fmt.Sprintf(
			"ℹ️ Вы уже подписаны на уведомления для кода %d\n\n"+
				"Вы получаете уведомления при обновлении списков каждые 5 минут.",
			uniqueCode,
		)
		buttons := []string{"Получить", "Отписаться"}
		commands := h.botHandler.SetKeyboardButtons(buttons, 2)
		h.botHandler.SendTextMessageWithKeyboardMarkup(message.Chat.ID, msg, commands)
		return
	}

	h.AddSubscription(message.Chat.ID, uniqueCode)

	// Запускаем мониторинг, если он еще не запущен
	h.StartMonitoring()

	msg := fmt.Sprintf(
		"✅ Вы подписались на уведомления для кода %d\n\n"+
			"Теперь вы будете получать уведомления при обновлении списков каждые 5 минут.",
		uniqueCode,
	)

	buttons := []string{"Получить", "Отписаться"}
	commands := h.botHandler.SetKeyboardButtons(buttons, 2)
	h.botHandler.SendTextMessageWithKeyboardMarkup(message.Chat.ID, msg, commands)
}

// handleUnsubscribeCommand обрабатывает команду отписки от уведомлений
func (h *MgsuHandler) handleUnsubscribeCommand(message *tgbotapi.Message) {
	// Проверяем, подписан ли пользователь
	if !h.IsSubscribed(message.Chat.ID) {
		msg := "ℹ️ Вы не подписаны на уведомления об обновлениях списков."
		buttons := []string{"Получить", "Подписаться"}
		commands := h.botHandler.SetKeyboardButtons(buttons, 2)
		h.botHandler.SendTextMessageWithKeyboardMarkup(message.Chat.ID, msg, commands)
		return
	}

	h.RemoveSubscription(message.Chat.ID)

	msg := "❌ Вы отписались от уведомлений об обновлениях списков."

	buttons := []string{"Получить", "Подписаться"}
	commands := h.botHandler.SetKeyboardButtons(buttons, 2)
	h.botHandler.SendTextMessageWithKeyboardMarkup(message.Chat.ID, msg, commands)
}

// ParseStudentPosition парсит страницу МГСУ и возвращает позицию студента
func (h *MgsuHandler) ParseStudentPosition(uniqueCode int) (*StudentInfo, error) {
	url := "https://mgsu.ru/2025/ks/bs/list.php?p=000000012_09.03.02_Informatsionnye_sistemy_i_tekhnologii_Ochnaya_Byudzhet_Obshchiy%20konkurs.html"

	// Загружаем страницу
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки страницы: %v", err)
	}
	defer resp.Body.Close()

	// Парсим HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга HTML: %v", err)
	}

	// Ищем количество бюджетных мест
	budgetPlaces := h.extractBudgetPlaces(doc)

	// Извлекаем дату и время создания списка
	creationDate, creationTime := h.extractCreationDateTime(doc)

	// Извлекаем направление обучения
	direction := h.extractDirection(doc)

	// Парсим таблицу и извлекаем данные студентов
	students, err := h.parseStudentTable(doc)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга таблицы: %v", err)
	}

	// Фильтруем студентов по высшему проходному приоритету (галочка в 6-м столбце "Это высший проходной приоритет")
	filteredStudents := h.filterByHighPassingPriority(students)

	// Ищем позицию студента с указанным кодом
	position, found := h.findStudentPosition(filteredStudents, uniqueCode)
	if !found {
		return nil, fmt.Errorf("студент с кодом %d не найден или не имеет высший проходной приоритет", uniqueCode)
	}

	// Вычисляем минимальный проходной балл
	minPassingScore := h.calculateMinPassingScore(filteredStudents, budgetPlaces)

	return &StudentInfo{
		BudgetPlaces:    budgetPlaces,
		Position:        fmt.Sprintf("%d/%d", position, budgetPlaces),
		MinPassingScore: minPassingScore,
		CreationDate:    creationDate,
		CreationTime:    creationTime,
		Direction:       direction,
	}, nil
}

// extractBudgetPlaces извлекает количество бюджетных мест из HTML
func (h *MgsuHandler) extractBudgetPlaces(doc *goquery.Document) int {
	// Ищем ячейку с текстом "Всего мест: X."
	budgetText := ""
	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.Contains(text, "Всего мест:") {
			budgetText = text
			return
		}
	})

	// Извлекаем число из текста формата "Всего мест: 107."
	if budgetText != "" {
		// Ищем паттерн "Всего мест: число."
		if strings.HasPrefix(budgetText, "Всего мест:") {
			// Убираем "Всего мест: " и "."
			numberPart := strings.TrimPrefix(budgetText, "Всего мест:")
			numberPart = strings.TrimSpace(numberPart)
			numberPart = strings.TrimSuffix(numberPart, ".")

			if num, err := strconv.Atoi(numberPart); err == nil {
				if num > 0 && num < 1000 { // разумные границы для количества мест
					return num
				}
			}
		}
	}

	// Если не нашли, возвращаем значение по умолчанию
	return 107
}

// extractCreationDateTime извлекает дату и время создания списка из HTML
func (h *MgsuHandler) extractCreationDateTime(doc *goquery.Document) (string, string) {
	// Ищем ячейку с текстом "Дата формирования - X. Время формирования - Y."
	var creationDate, creationTime string

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.Contains(text, "Дата формирования") && strings.Contains(text, "Время формирования") {
			// Парсим строку формата "Дата формирования - 31.07.2025. Время формирования - 10:01:01."

			// Ищем дату после "Дата формирования - "
			if dateStart := strings.Index(text, "Дата формирования - "); dateStart != -1 {
				dateStart += len("Дата формирования - ")
				// Ищем следующую точку после даты
				if dateEnd := strings.Index(text[dateStart:], ". Время формирования"); dateEnd != -1 {
					creationDate = strings.TrimSpace(text[dateStart : dateStart+dateEnd])
				}
			}

			// Ищем время после "Время формирования - "
			if timeStart := strings.Index(text, "Время формирования - "); timeStart != -1 {
				timeStart += len("Время формирования - ")
				// Ищем следующую точку после времени
				if timeEnd := strings.Index(text[timeStart:], "."); timeEnd != -1 {
					creationTime = strings.TrimSpace(text[timeStart : timeStart+timeEnd])
				} else {
					// Если точки нет, берем до конца строки
					creationTime = strings.TrimSpace(text[timeStart:])
				}
			}
			return
		}
	})

	return creationDate, creationTime
}

// extractDirection извлекает направление обучения из HTML
func (h *MgsuHandler) extractDirection(doc *goquery.Document) string {
	// Ищем ячейку с текстом "Конкурсная группа - X"
	direction := ""

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.Contains(text, "Конкурсная группа") {
			// Извлекаем направление из строки формата "Конкурсная группа - 09.03.02_Информационные_системы_и_технологии_Очная_Бюджет_Общий конкурс"
			direction = strings.TrimPrefix(text, "Конкурсная группа - ")
			direction = strings.TrimSpace(direction)
			// Заменяем подчеркивания на пробелы для лучшей читаемости
			direction = strings.ReplaceAll(direction, "_", " ")
			return
		}
	})

	return direction
}

// formatDate форматирует дату из формата DD.MM.YYYY в читаемый вид
func (h *MgsuHandler) formatDate(dateStr string) string {
	// Просто возвращаем дату как есть
	return dateStr
}

// parseStudentTable парсит таблицу студентов
func (h *MgsuHandler) parseStudentTable(doc *goquery.Document) ([]StudentEntry, error) {
	var students []StudentEntry

	// Ищем таблицу с данными
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		// Проверяем, что это нужная нам таблица по заголовкам
		headers := table.Find("tr.header-row th")
		if headers.Length() < 16 {
			return
		}

		// Парсим строки данных
		table.Find("tr.data-row").Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() >= 16 {
				student := StudentEntry{
					Number:                cells.Eq(0).Text(),                     // №
					UniqueCode:            strings.TrimSpace(cells.Eq(1).Text()),  // Уникальный код
					Priority:              strings.TrimSpace(cells.Eq(2).Text()),  // Приоритет
					AdmissionConsent:      strings.TrimSpace(cells.Eq(3).Text()),  // Согласие на зачисление
					HighPassingPriority:   strings.TrimSpace(cells.Eq(4).Text()),  // Высший проходной приоритет
					IsHighPassingPriority: strings.TrimSpace(cells.Eq(5).Text()),  // Это высший проходной приоритет
					MainHighPriority:      strings.TrimSpace(cells.Eq(6).Text()),  // Основной высший приоритет
					TotalScore:            strings.TrimSpace(cells.Eq(7).Text()),  // Сумма баллов
					SubjectScore:          strings.TrimSpace(cells.Eq(8).Text()),  // Сумма по предметам
					Math:                  strings.TrimSpace(cells.Eq(9).Text()),  // Матем / ЧиИГ
					IT:                    strings.TrimSpace(cells.Eq(10).Text()), // ИиИКТ / Физика / БезопЖизнедеят
					Russian:               strings.TrimSpace(cells.Eq(11).Text()), // РусЯз
					GeneralAchievements:   strings.TrimSpace(cells.Eq(12).Text()), // Общие ИД
					BVIBasis:              strings.TrimSpace(cells.Eq(13).Text()), // Основание БВИ
					PPR9:                  strings.TrimSpace(cells.Eq(14).Text()), // ППР (ч.9 с. 71 273-ФЗ)
					PPR10:                 strings.TrimSpace(cells.Eq(15).Text()), // ППР (ч.10 с. 71 273-ФЗ)
					IsMainHighPriority:    strings.TrimSpace(cells.Eq(16).Text()), // Номер предложения (используем поле IsMainHighPriority)
					// Остальные столбцы пока не используем:
					// cells.Eq(17) - Размещено на РВР
					// cells.Eq(18) - ID заказчика (нет на РВР)
					// cells.Eq(19) - Целевые ИД
				}
				students = append(students, student)
			}
		})
	})

	if len(students) == 0 {
		return nil, fmt.Errorf("таблица студентов не найдена")
	}

	return students, nil
}

// filterByHighPassingPriority фильтрует студентов по наличию галочки в колонке "Это высший проходной приоритет"
func (h *MgsuHandler) filterByHighPassingPriority(students []StudentEntry) []StudentEntry {
	var filtered []StudentEntry

	for _, student := range students {
		// Проверяем наличие галочки (✓) в колонке "Это высший проходной приоритет"
		if strings.Contains(student.IsHighPassingPriority, "✓") {
			filtered = append(filtered, student)
		}
	}

	return filtered
}

// findStudentPosition находит позицию студента в отфильтрованном списке
func (h *MgsuHandler) findStudentPosition(students []StudentEntry, uniqueCode int) (int, bool) {
	codeStr := strconv.Itoa(uniqueCode)

	for i, student := range students {
		if student.UniqueCode == codeStr {
			return i + 1, true // позиция начинается с 1
		}
	}

	return 0, false
}

// calculateMinPassingScore вычисляет минимальный проходной балл
func (h *MgsuHandler) calculateMinPassingScore(students []StudentEntry, budgetPlaces int) int {
	if len(students) < budgetPlaces {
		// Если студентов меньше чем мест, берем балл последнего
		if len(students) > 0 {
			if score, err := strconv.Atoi(students[len(students)-1].TotalScore); err == nil {
				return score
			}
		}
		return 0
	}

	// Берем балл студента на последнем проходном месте
	if score, err := strconv.Atoi(students[budgetPlaces-1].TotalScore); err == nil {
		return score
	}

	return 0
}

// StartMonitoring запускает мониторинг изменений в списках
func (h *MgsuHandler) StartMonitoring() {
	h.mutex.Lock()
	if h.monitoringActive {
		h.mutex.Unlock()
		return
	}
	h.monitoringActive = true
	h.mutex.Unlock()

	go h.monitoringLoop()
}

// StopMonitoring останавливает мониторинг
func (h *MgsuHandler) StopMonitoring() {
	h.mutex.Lock()
	h.monitoringActive = false
	h.mutex.Unlock()
}

// AddSubscription добавляет пользователя в список уведомлений
func (h *MgsuHandler) AddSubscription(chatID int64, uniqueCode int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.subscribedUsers[chatID] = uniqueCode
}

// RemoveSubscription удаляет пользователя из списка уведомлений
func (h *MgsuHandler) RemoveSubscription(chatID int64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.subscribedUsers, chatID)
}

// IsSubscribed проверяет, подписан ли пользователь на уведомления
func (h *MgsuHandler) IsSubscribed(chatID int64) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	_, exists := h.subscribedUsers[chatID]
	return exists
}

// monitoringLoop основной цикл мониторинга
func (h *MgsuHandler) monitoringLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mutex.RLock()
		active := h.monitoringActive
		h.mutex.RUnlock()

		if !active {
			return
		}

		h.checkForUpdates()
	}
}

// checkForUpdates проверяет обновления списка
func (h *MgsuHandler) checkForUpdates() {
	url := "https://mgsu.ru/2025/ks/bs/list.php?p=000000012_09.03.02_Informatsionnye_sistemy_i_tekhnologii_Ochnaya_Byudzhet_Obshchiy%20konkurs.html"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Ошибка при проверке обновлений: %v\n", err)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Ошибка парсинга при проверке обновлений: %v\n", err)
		return
	}

	// Извлекаем дату и время создания
	creationDate, creationTime := h.extractCreationDateTime(doc)
	currentDateTime := fmt.Sprintf("%s %s", creationDate, creationTime)

	h.mutex.Lock()
	lastDateTime := h.lastCreationDateTime
	h.mutex.Unlock()

	// Если время изменилось, отправляем уведомления
	if lastDateTime != "" && lastDateTime != currentDateTime {
		fmt.Printf("Обнаружено обновление списка: %s -> %s\n", lastDateTime, currentDateTime)
		h.sendUpdateNotifications()
	}

	// Обновляем последнее время
	h.mutex.Lock()
	h.lastCreationDateTime = currentDateTime
	h.mutex.Unlock()
}

// sendUpdateNotifications отправляет уведомления всем подписанным пользователям
func (h *MgsuHandler) sendUpdateNotifications() {
	h.mutex.RLock()
	subscribers := make(map[int64]int)
	for chatID, uniqueCode := range h.subscribedUsers {
		subscribers[chatID] = uniqueCode
	}
	h.mutex.RUnlock()

	for chatID, uniqueCode := range subscribers {
		go h.sendNotificationToUser(chatID, uniqueCode)
	}
}

// sendNotificationToUser отправляет уведомление конкретному пользователю
func (h *MgsuHandler) sendNotificationToUser(chatID int64, uniqueCode int) {
	studentInfo, err := h.ParseStudentPosition(uniqueCode)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Ошибка при получении обновленной информации для кода %d: %v", uniqueCode, err)
		h.botHandler.SendTextMessage(chatID, errorMsg)
		return
	}

	msg := fmt.Sprintf(
		"🔔 ОБНОВЛЕНИЕ СПИСКА!\n\n"+
			"Информация о студенте с кодом %d:\n"+
			"🎯 Позиция: %s\n"+
			"📚 Количество бюджетных мест: %d\n"+
			"📊 Минимальный проходной балл: %d\n"+
			"📅 Дата создания: %s\n"+
			"⏰ Время создания: %s\n"+
			"🎓 Направление: %s",
		uniqueCode,
		studentInfo.Position,
		studentInfo.BudgetPlaces,
		studentInfo.MinPassingScore,
		h.formatDate(studentInfo.CreationDate),
		studentInfo.CreationTime,
		studentInfo.Direction,
	)

	h.botHandler.SendTextMessage(chatID, msg)
}
