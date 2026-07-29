package exports

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
	"smart-livestock-backend/models"
)

type ExcelCowData struct {
	Cow             models.Cow
	Weighings       []models.WeightRecord
	LastWeight      float64
	LastADG         float64
	PredictedWeight float64
	StatusDSS       string
}

func GenerateExcelReport(data []ExcelCowData, breedFilter string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// ============================================================
	// STYLES
	// ============================================================
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1A1A1A"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleDataCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleDataRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	styleLayak, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "0F5132"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D1E7DD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleEvaluasi, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "664D03"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFF3CD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleTidakLayak, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "842029"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F8D7DA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	// ============================================================
	// SHEET 1: Ringkasan & DSS
	// ============================================================
	sheetSummary := "Ringkasan & DSS"
	f.SetSheetName("Sheet1", sheetSummary)

	headers := []string{
		"No",
		"Kode RFID / Tag",
		"Nama Sapi",
		"Rumpun Sapi",
		"Jenis Kelamin",
		"Pemilik / Kandang",
		"Jml Timbang",
		"Waktu Timbang 1 (Tgl & Jam)",
		"Bobot Timbang 1 (KG)",
		"Waktu Timbang 2 (Tgl & Jam)",
		"Bobot Timbang 2 (KG)",
		"Waktu Timbang 3 (Tgl & Jam)",
		"Bobot Timbang 3 (KG)",
		"Rata-rata ADG (KG/Hari)",
		"Prediksi Bobot (+30 Hari)",
		"Status Rekomendasi DSS",
	}

	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetSummary, cell, h)
		f.SetCellStyle(sheetSummary, cell, cell, styleHeader)
	}
	f.SetRowHeight(sheetSummary, 1, 32)

	for idx, item := range data {
		rowNum := idx + 2
		f.SetRowHeight(sheetSummary, rowNum, 22)

		tgl1, w1 := "-", "-"
		tgl2, w2 := "-", "-"
		tgl3, w3 := "-", "-"

		if len(item.Weighings) > 0 {
			tgl1 = item.Weighings[0].MeasurementDate.Format("02-01-2006 15:04")
			w1 = fmt.Sprintf("%.2f", item.Weighings[0].Weight)
		}
		if len(item.Weighings) > 1 {
			tgl2 = item.Weighings[1].MeasurementDate.Format("02-01-2006 15:04")
			w2 = fmt.Sprintf("%.2f", item.Weighings[1].Weight)
		}
		if len(item.Weighings) > 2 {
			tgl3 = item.Weighings[2].MeasurementDate.Format("02-01-2006 15:04")
			w3 = fmt.Sprintf("%.2f", item.Weighings[2].Weight)
		}

		owner := "-"
		if item.Cow.Owner != nil && *item.Cow.Owner != "" {
			owner = *item.Cow.Owner
		}

		statusLabel := item.StatusDSS
		if statusLabel == "" {
			statusLabel = "Belum Cukup Data (< 3x Timbang)"
		}

		predStr := "-"
		if item.PredictedWeight > 0 {
			predStr = fmt.Sprintf("%.2f", item.PredictedWeight)
		}

		f.SetCellValue(sheetSummary, fmt.Sprintf("A%d", rowNum), idx+1)
		f.SetCellValue(sheetSummary, fmt.Sprintf("B%d", rowNum), item.Cow.CowCode)
		f.SetCellValue(sheetSummary, fmt.Sprintf("C%d", rowNum), item.Cow.Name)
		f.SetCellValue(sheetSummary, fmt.Sprintf("D%d", rowNum), item.Cow.Breed)
		f.SetCellValue(sheetSummary, fmt.Sprintf("E%d", rowNum), item.Cow.Gender)
		f.SetCellValue(sheetSummary, fmt.Sprintf("F%d", rowNum), owner)
		f.SetCellValue(sheetSummary, fmt.Sprintf("G%d", rowNum), len(item.Weighings))
		f.SetCellValue(sheetSummary, fmt.Sprintf("H%d", rowNum), tgl1)
		f.SetCellValue(sheetSummary, fmt.Sprintf("I%d", rowNum), w1)
		f.SetCellValue(sheetSummary, fmt.Sprintf("J%d", rowNum), tgl2)
		f.SetCellValue(sheetSummary, fmt.Sprintf("K%d", rowNum), w2)
		f.SetCellValue(sheetSummary, fmt.Sprintf("L%d", rowNum), tgl3)
		f.SetCellValue(sheetSummary, fmt.Sprintf("M%d", rowNum), w3)
		f.SetCellValue(sheetSummary, fmt.Sprintf("N%d", rowNum), fmt.Sprintf("%.2f", item.LastADG))
		f.SetCellValue(sheetSummary, fmt.Sprintf("O%d", rowNum), predStr)
		f.SetCellValue(sheetSummary, fmt.Sprintf("P%d", rowNum), statusLabel)

		f.SetCellStyle(sheetSummary, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), styleDataCenter)
		f.SetCellStyle(sheetSummary, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("H%d", rowNum), styleDataCenter)
		f.SetCellStyle(sheetSummary, fmt.Sprintf("I%d", rowNum), fmt.Sprintf("I%d", rowNum), styleDataRight)
		f.SetCellStyle(sheetSummary, fmt.Sprintf("J%d", rowNum), fmt.Sprintf("J%d", rowNum), styleDataCenter)
		f.SetCellStyle(sheetSummary, fmt.Sprintf("K%d", rowNum), fmt.Sprintf("K%d", rowNum), styleDataRight)
		f.SetCellStyle(sheetSummary, fmt.Sprintf("L%d", rowNum), fmt.Sprintf("L%d", rowNum), styleDataCenter)
		f.SetCellStyle(sheetSummary, fmt.Sprintf("M%d", rowNum), fmt.Sprintf("O%d", rowNum), styleDataRight)

		cellStatus := fmt.Sprintf("P%d", rowNum)
		switch statusLabel {
		case "LAYAK DIPERTAHANKAN":
			f.SetCellStyle(sheetSummary, cellStatus, cellStatus, styleLayak)
		case "PERLU EVALUASI":
			f.SetCellStyle(sheetSummary, cellStatus, cellStatus, styleEvaluasi)
		case "TIDAK LAYAK DIPERTAHANKAN":
			f.SetCellStyle(sheetSummary, cellStatus, cellStatus, styleTidakLayak)
		default:
			f.SetCellStyle(sheetSummary, cellStatus, cellStatus, styleDataCenter)
		}
	}

	f.SetColWidth(sheetSummary, "A", "A", 6)
	f.SetColWidth(sheetSummary, "B", "B", 18)
	f.SetColWidth(sheetSummary, "C", "D", 20)
	f.SetColWidth(sheetSummary, "E", "G", 15)
	f.SetColWidth(sheetSummary, "H", "H", 24)
	f.SetColWidth(sheetSummary, "I", "I", 18)
	f.SetColWidth(sheetSummary, "J", "J", 24)
	f.SetColWidth(sheetSummary, "K", "K", 18)
	f.SetColWidth(sheetSummary, "L", "L", 24)
	f.SetColWidth(sheetSummary, "M", "M", 18)
	f.SetColWidth(sheetSummary, "N", "O", 22)
	f.SetColWidth(sheetSummary, "P", "P", 28)

	// ============================================================
	// SHEET 2: Log Timbangan Detail
	// ============================================================
	sheetDetail := "Log Timbangan Detail"
	f.NewSheet(sheetDetail)
	detailHeaders := []string{
		"No",
		"Kode RFID / Tag",
		"Nama Sapi",
		"Rumpun Sapi",
		"Bobot Timbang (KG)",
		"ADG Harian (KG/Hari)",
		"Tanggal & Waktu Timbang",
		"ID Perangkat Timbangan",
	}

	for colIdx, h := range detailHeaders {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetDetail, cell, h)
		f.SetCellStyle(sheetDetail, cell, cell, styleHeader)
	}
	f.SetRowHeight(sheetDetail, 1, 32)

	detailRow := 2
	for _, item := range data {
		for _, w := range item.Weighings {
			f.SetRowHeight(sheetDetail, detailRow, 20)
			adgVal := 0.0
			if w.ADG != nil {
				adgVal = *w.ADG
			}
			devID := "SCALE-ESP32-01"
			if w.DeviceID != nil && *w.DeviceID != "" {
				devID = *w.DeviceID
			}

			f.SetCellValue(sheetDetail, fmt.Sprintf("A%d", detailRow), detailRow-1)
			f.SetCellValue(sheetDetail, fmt.Sprintf("B%d", detailRow), item.Cow.CowCode)
			f.SetCellValue(sheetDetail, fmt.Sprintf("C%d", detailRow), item.Cow.Name)
			f.SetCellValue(sheetDetail, fmt.Sprintf("D%d", detailRow), item.Cow.Breed)
			f.SetCellValue(sheetDetail, fmt.Sprintf("E%d", detailRow), w.Weight)
			f.SetCellValue(sheetDetail, fmt.Sprintf("F%d", detailRow), fmt.Sprintf("%.2f", adgVal))
			f.SetCellValue(sheetDetail, fmt.Sprintf("G%d", detailRow), w.MeasurementDate.Format("02-01-2006 15:04:05"))
			f.SetCellValue(sheetDetail, fmt.Sprintf("H%d", detailRow), devID)

			f.SetCellStyle(sheetDetail, fmt.Sprintf("A%d", detailRow), fmt.Sprintf("B%d", detailRow), styleDataCenter)
			f.SetCellStyle(sheetDetail, fmt.Sprintf("E%d", detailRow), fmt.Sprintf("F%d", detailRow), styleDataRight)
			f.SetCellStyle(sheetDetail, fmt.Sprintf("G%d", detailRow), fmt.Sprintf("H%d", detailRow), styleDataCenter)

			detailRow++
		}
	}

	f.SetColWidth(sheetDetail, "A", "A", 6)
	f.SetColWidth(sheetDetail, "B", "B", 18)
	f.SetColWidth(sheetDetail, "C", "D", 20)
	f.SetColWidth(sheetDetail, "E", "F", 20)
	f.SetColWidth(sheetDetail, "G", "H", 25)

	// ============================================================
	// SHEET 3: Data Sumber Grafik
	// Setiap sapi menempati 2 kolom: [Tanggal | Bobot]
	// Kolom layout: cow0→A,B | cow1→D,E | cow2→G,H | ... (gap 1 kolom antar sapi)
	// ============================================================
	sheetChartData := "Data Sumber Grafik"
	f.NewSheet(sheetChartData)

	styleChartHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 9},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2563EB"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	const colsPerCow = 3 // 2 data cols + 1 gap
	type cowChartRange struct {
		dateCol   string // e.g. "A"
		weightCol string // e.g. "B"
		startRow  int
		endRow    int
		name      string
	}
	cowRanges := make([]cowChartRange, 0, len(data))

	for cowIdx, item := range data {
		if len(item.Weighings) == 0 {
			continue
		}

		baseCol := cowIdx*colsPerCow + 1 // 1-indexed column number
		dateColName, _ := excelize.ColumnNumberToName(baseCol)
		weightColName, _ := excelize.ColumnNumberToName(baseCol + 1)

		// Header row
		headerDateCell := fmt.Sprintf("%s1", dateColName)
		headerWeightCell := fmt.Sprintf("%s1", weightColName)
		f.SetCellValue(sheetChartData, headerDateCell, fmt.Sprintf("%s - Tanggal", item.Cow.Name))
		f.SetCellValue(sheetChartData, headerWeightCell, fmt.Sprintf("%s - Bobot (Kg)", item.Cow.Name))
		f.SetCellStyle(sheetChartData, headerDateCell, headerWeightCell, styleChartHeader)
		f.SetColWidth(sheetChartData, dateColName, dateColName, 14)
		f.SetColWidth(sheetChartData, weightColName, weightColName, 14)

		// Data rows
		startRow := 2
		for rowIdx, w := range item.Weighings {
			row := startRow + rowIdx
			dateCell := fmt.Sprintf("%s%d", dateColName, row)
			weightCell := fmt.Sprintf("%s%d", weightColName, row)
			f.SetCellValue(sheetChartData, dateCell, w.MeasurementDate.Format("02/01/06"))
			f.SetCellValue(sheetChartData, weightCell, w.Weight)
		}

		endRow := startRow + len(item.Weighings) - 1
		cowRanges = append(cowRanges, cowChartRange{
			dateCol:   dateColName,
			weightCol: weightColName,
			startRow:  startRow,
			endRow:    endRow,
			name:      item.Cow.Name,
		})
	}

	// ============================================================
	// SHEET 4: Grafik Pertumbuhan
	// Satu line chart per sapi, disusun vertikal (satu per baris)
	// Setiap chart: lebar ~500px (8 col), tinggi ~14 baris
	// ============================================================
	sheetChart := "Grafik Pertumbuhan"
	f.NewSheet(sheetChart)

	const chartHeightRows = 15
	const chartWidthCols = 10

	for i, cr := range cowRanges {
		// Posisi chart: setiap chart menempati ~chartHeightRows baris
		anchorRow := i*chartHeightRows + 1
		anchorCell, _ := excelize.CoordinatesToCellName(1, anchorRow)

		catRange := fmt.Sprintf("'%s'!$%s$%d:$%s$%d",
			sheetChartData, cr.dateCol, cr.startRow, cr.dateCol, cr.endRow)
		valRange := fmt.Sprintf("'%s'!$%s$%d:$%s$%d",
			sheetChartData, cr.weightCol, cr.startRow, cr.weightCol, cr.endRow)
		seriesName := fmt.Sprintf("'%s'!$%s$1", sheetChartData, cr.weightCol)

		_ = chartWidthCols // reserved for future multi-column layout

		err := f.AddChart(sheetChart, anchorCell, &excelize.Chart{
			Type: excelize.Line,
			Series: []excelize.ChartSeries{
				{
					Name:       seriesName,
					Categories: catRange,
					Values:     valRange,
					Line: excelize.LineOptions{
						Smooth: true,
						Width:  2.5,
					},
					Marker: excelize.ChartMarker{
						Symbol: "circle",
						Size:   6,
					},
				},
			},
			Title: excelize.ChartTitle{
				Paragraph: []excelize.RichTextRun{
					{Text: fmt.Sprintf("Tren Pertumbuhan Bobot — %s", cr.name)},
				},
			},
			Legend: excelize.ChartLegend{
				Position: "bottom",
			},
			PlotArea: excelize.ChartPlotArea{
				ShowVal:     true,
				ShowCatName: false,
			},
			Dimension: excelize.ChartDimension{
				Width:  500,
				Height: 300,
			},
		})
		if err != nil {
			// Jika gagal buat grafik, lanjutkan saja (tidak batalkan seluruh export)
			_ = err
		}
	}

	// ============================================================
	// Save to Buffer
	// ============================================================
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
