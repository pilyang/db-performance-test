package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	port     int
	name     string
	testName string
	queries  []QueryTest
}

type QueryTest struct {
	name  string
	query string
}

type TestResult struct {
	rowCount      int
	executionTime time.Duration
	planningTime  time.Duration
	totalTime     time.Duration
	queryPlan     []string
}

type TestSummary struct {
	avgExecutionTime time.Duration
	avgPlanningTime  time.Duration
	avgTotalTime     time.Duration
	avgRowCount      int
	minExecutionTime time.Duration
	maxExecutionTime time.Duration
	runs             []TestResult
}

func main() {
	// 데이터베이스 설정
	configs := []DBConfig{
		{
			port:     5433,
			name:     "TEXT[] Array",
			testName: "text_array",
			queries: []QueryTest{
				{
					name: "Single Tag Search",
					query: `
						SELECT * FROM articles_with_text_array
						WHERE tag_values @> ARRAY['javascript']`,
				},
				{
					name: "Multiple Tags AND",
					query: `
						SELECT * FROM articles_with_text_array
						WHERE tag_values @> ARRAY['javascript', 'jeact']`,
				},
				{
					name: "Multiple Tags OR",
					query: `
						SELECT * FROM articles_with_text_array
						WHERE tag_values && ARRAY['javascript', 'python']`,
				},
			},
		},
		{
			port:     5434,
			name:     "INTEGER[] Array",
			testName: "int_array",
			queries: []QueryTest{
				{
					name: "Single Tag Search",
					query: `
						SELECT a.* 
						FROM articles_with_int_array a
						WHERE tag_ids @> ARRAY[(SELECT id FROM tags WHERE value = 'javascript')]`,
				},
				{
					name: "Multiple Tags AND",
					query: `
						SELECT a.* 
						FROM articles_with_int_array a
						WHERE tag_ids @> ARRAY[
							(SELECT id FROM tags WHERE value = 'javascript'),
							(SELECT id FROM tags WHERE value = 'react')
						]`,
				},
				{
					name: "Multiple Tags OR",
					query: `
						SELECT a.* 
						FROM articles_with_int_array a
						WHERE tag_ids && ARRAY[
							(SELECT id FROM tags WHERE value = 'javascript'),
							(SELECT id FROM tags WHERE value = 'python')
						]`,
				},
			},
		},
		{
			port:     5432,
			name:     "Relational Table",
			testName: "relational",
			queries: []QueryTest{
				{
					name: "Single Tag Search",
					query: `
						SELECT DISTINCT a.* 
						FROM articles a
						JOIN article_tags at ON a.id = at.article_id
						JOIN tags t ON t.id = at.tag_id
						WHERE t.value = 'javascript'`,
				},
				{
					name: "Multiple Tags AND",
					query: `
						SELECT a.* 
						FROM articles a
						WHERE EXISTS (
							SELECT 1 FROM article_tags at1 
							JOIN tags t1 ON t1.id = at1.tag_id 
							WHERE at1.article_id = a.id AND t1.value = 'javascript'
						)
						AND EXISTS (
							SELECT 1 FROM article_tags at2 
							JOIN tags t2 ON t2.id = at2.tag_id 
							WHERE at2.article_id = a.id AND t2.value = 'react'
						)`,
				},
				{
					name: "Multiple Tags OR",
					query: `
						SELECT DISTINCT a.* 
						FROM articles a
						JOIN article_tags at ON a.id = at.article_id
						JOIN tags t ON t.id = at.tag_id
						WHERE t.value IN ('javascript', 'python')`,
				},
			},
		},
	}

	// 데이터베이스 연결 및 워밍업
	dbs := initializeDatabases(configs)
	defer closeDatabases(dbs)

	// 테스트 실행
	for _, config := range configs {
		if db, ok := dbs[config.port]; ok {
			runTestsForDB(config, db)
		}
	}
}

func initializeDatabases(configs []DBConfig) map[int]*sql.DB {
	dbs := make(map[int]*sql.DB)
	for _, config := range configs {
		db, err := connectDB(config.port)
		if err != nil {
			log.Printf("Failed to connect to database %s: %v\n", config.name, err)
			continue
		}

		// Connection pool 설정
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)

		// 워밍업
		warmupDB(db)
		dbs[config.port] = db
	}
	return dbs
}

func closeDatabases(dbs map[int]*sql.DB) {
	for _, db := range dbs {
		db.Close()
	}
}

func connectDB(port int) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=localhost port=%d user=test password=mypassword dbname=test_db sslmode=disable",
		port,
	)
	return sql.Open("postgres", connStr)
}

func warmupDB(db *sql.DB) {
	for i := 0; i < 5; i++ {
		rows, err := db.Query("SELECT 1")
		if err == nil {
			rows.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func executeQuery(db *sql.DB, query string) (TestResult, error) {
	explainQuery := "EXPLAIN (ANALYZE, BUFFERS) " + query
	start := time.Now()

	rows, err := db.Query(explainQuery)
	if err != nil {
		return TestResult{}, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()
	totalTime := time.Since(start)

	var planningTime, executionTime float64
	var rowCount int
	var queryPlan []string
	var foundRows bool

	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return TestResult{}, fmt.Errorf("scanning row error: %w", err)
		}
		queryPlan = append(queryPlan, line)

		// Planning Time 추출
		if strings.Contains(line, "Planning Time") {
			fmt.Sscanf(line, "Planning Time: %f ms", &planningTime)
		}
		// Execution Time 추출
		if strings.Contains(line, "Execution Time") {
			fmt.Sscanf(line, "Execution Time: %f ms", &executionTime)
		}

		// 실제 결과 행 수 추출
		// 최상위 노드의 actual rows 값을 찾음
		if !foundRows && strings.Contains(line, "actual time=") {
			rowsStr := line[strings.Index(line, "rows=")+5:]
			rowsStr = rowsStr[:strings.Index(rowsStr, " ")]
			if rows, err := strconv.Atoi(rowsStr); err == nil {
				rowCount = rows
				foundRows = true
			}
		}
	}

	if err = rows.Err(); err != nil {
		return TestResult{}, fmt.Errorf("rows iteration error: %w", err)
	}

	return TestResult{
		rowCount:      rowCount,
		executionTime: time.Duration(executionTime * float64(time.Millisecond)),
		planningTime:  time.Duration(planningTime * float64(time.Millisecond)),
		totalTime:     totalTime,
		queryPlan:     queryPlan,
	}, nil
}

func runTestsForDB(config DBConfig, db *sql.DB) {
	fmt.Printf("\n=== Testing %s (Port: %d) ===\n", config.name, config.port)
	fmt.Println(strings.Repeat("=", 50))

	summaries := make(map[string]*TestSummary)

	// 각 쿼리에 대한 테스트 실행
	for _, qt := range config.queries {
		summary := runQueryTest(db, qt)
		summaries[qt.name] = summary

		printQuerySummary(qt.name, summary)
		time.Sleep(500 * time.Millisecond) // 쿼리 간 간격
	}

	// 전체 결과 요약 출력
	printDBSummary(config.name, summaries)
}

func runQueryTest(db *sql.DB, qt QueryTest) *TestSummary {
	fmt.Printf("\nExecuting: %s\n", qt.name)
	fmt.Println(strings.Repeat("-", 30))

	summary := &TestSummary{
		minExecutionTime: time.Duration(1<<63 - 1),
		runs:             make([]TestResult, 0),
	}

	// 워밍업 실행
	_, _ = executeQuery(db, qt.query)
	time.Sleep(200 * time.Millisecond)

	// 실제 테스트 실행 (10회)
	for i := 0; i < 10; i++ {
		result, err := executeQuery(db, qt.query)
		if err != nil {
			log.Printf("Error executing query: %v\n", err)
			continue
		}

		summary.runs = append(summary.runs, result)
		summary.avgExecutionTime += result.executionTime
		summary.avgPlanningTime += result.planningTime
		summary.avgTotalTime += result.totalTime
		summary.avgRowCount += result.rowCount

		if result.executionTime < summary.minExecutionTime {
			summary.minExecutionTime = result.executionTime
		}
		if result.executionTime > summary.maxExecutionTime {
			summary.maxExecutionTime = result.executionTime
		}

		fmt.Printf("Run %d:\n", i+1)
		fmt.Printf("  Execution Time: %v\n", result.executionTime)
		fmt.Printf("  Planning Time: %v\n", result.planningTime)
		fmt.Printf("  Total Time: %v\n", result.totalTime)
		fmt.Printf("  Rows: %d\n", result.rowCount)

		time.Sleep(50 * time.Millisecond)
	}

	// 평균 계산
	count := len(summary.runs)
	if count > 0 {
		summary.avgExecutionTime /= time.Duration(count)
		summary.avgPlanningTime /= time.Duration(count)
		summary.avgTotalTime /= time.Duration(count)
		summary.avgRowCount /= count
	}

	return summary
}

func printQuerySummary(queryName string, summary *TestSummary) {
	fmt.Printf("\nSummary for %s:\n", queryName)
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("Average Execution Time: %v\n", summary.avgExecutionTime)
	fmt.Printf("Average Planning Time: %v\n", summary.avgPlanningTime)
	fmt.Printf("Average Total Time: %v\n", summary.avgTotalTime)
	fmt.Printf("Average Row Count: %d\n", summary.avgRowCount)
	fmt.Printf("Min Execution Time: %v\n", summary.minExecutionTime)
	fmt.Printf("Max Execution Time: %v\n", summary.maxExecutionTime)

	// 실행 계획 출력 (첫 번째 실행의 계획만)
	if len(summary.runs) > 0 {
		fmt.Println("\nQuery Plan (first run):")
		for _, line := range summary.runs[0].queryPlan {
			fmt.Println(line)
		}
	}
}

func printDBSummary(dbName string, summaries map[string]*TestSummary) {
	fmt.Printf("\n=== Overall Summary for %s ===\n", dbName)
	fmt.Println(strings.Repeat("=", 50))

	for queryName, summary := range summaries {
		fmt.Printf("\n%s:\n", queryName)
		fmt.Printf("  Avg Execution Time: %v\n", summary.avgExecutionTime)
		fmt.Printf("  Avg Planning Time: %v\n", summary.avgPlanningTime)
		fmt.Printf("  Execution Time Range: %v - %v\n",
			summary.minExecutionTime, summary.maxExecutionTime)
	}
}
