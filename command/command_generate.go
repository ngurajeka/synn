package command

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/ngurajeka/synn/shared"

	"github.com/mitchellh/cli"
	"github.com/tealeg/xlsx"
	"go.uber.org/zap"
)

type GenerateCmd struct {
	ui     cli.Ui
	logger *zap.Logger
}

func NewGenerateCmd(ui cli.Ui, logger *zap.Logger) *GenerateCmd {
	return &GenerateCmd{ui, logger}
}

func (c *GenerateCmd) Help() string {
	helpText := `
		Usage: accent dry-run [options] ...
  		Checking Certificates.
	`
	return strings.TrimSpace(helpText)
}

func (c *GenerateCmd) Synopsis() string {
	return "Running the dry run process"
}

type Header struct {
	Name     string
	CellType int
}

const (
	CellTypeString = iota
	CellTypeNumber
	CellTypeDate
)

func (c *GenerateCmd) Run(args []string) int {
	template, _ := pongo2.FromFile("template.txt")
	file, exist := shared.GetKey("data", args)
	if !exist {
		c.logger.Info("data argument not found", zap.Strings("args", args))
		return 1
	}

	// reading xlsx file
	excelFile, err := xlsx.OpenFile(file)
	if err != nil {
		c.logger.Info("reading file excel failed", zap.String("file", file), zap.Error(err))
		return 1
	}

	// parsing xlsx
	for _, sheet := range excelFile.Sheets {
		var (
			headers []Header
			rows    [][]string
		)
		for i, row := range sheet.Rows {
			var rowData []string
			for j, cell := range row.Cells {
				if i == 0 {
					headers = append(headers, Header{Name: shared.NormalizeKey(cell.String())})
					continue
				}
				header := headers[j]
				switch cell.Type() {
				case xlsx.CellTypeDate:
					header.CellType = CellTypeDate
					t, err := cell.GetTime(false)
					if err != nil {
						c.logger.Error("parsing date error",
							zap.String("sheet", sheet.Name),
							zap.Int("row", i),
							zap.Int("column index", j),
							zap.String("column", header.Name))
						continue
					}
					rowData = append(rowData, t.String())
				case xlsx.CellTypeFormula:
					header.CellType = CellTypeString
					rowData = append(rowData, cell.Value)
				case xlsx.CellTypeString:
					header.CellType = CellTypeString
					rowData = append(rowData, cell.String())
				case xlsx.CellTypeNumeric:
					if cell.NumFmt == "dd\\-mm\\-yyyy\\ hh\\:mm\\:ss;@" {
						header.CellType = CellTypeDate
						t, err := cell.GetTime(false)
						if err != nil {
							c.logger.Error("parsing date error",
								zap.String("sheet", sheet.Name),
								zap.Int("row", i),
								zap.Int("column index", j),
								zap.String("column", header.Name))
							continue
						}
						rowData = append(rowData, t.Format("2006-01-02 15:04:05"))
						continue
					}
					header.CellType = CellTypeNumber
					v, err := cell.Int64()
					if err != nil {
						c.logger.Error("parsing integer error",
							zap.String("sheet", sheet.Name),
							zap.Int("row", i),
							zap.Int("column index", j),
							zap.String("numFmt", cell.NumFmt),
							zap.String("column", header.Name))
						continue
					}
					rowData = append(rowData, strconv.FormatInt(v, 10))
				default:
					c.logger.Error("unknown cell type",
						zap.String("sheet", sheet.Name),
						zap.Int("row", i),
						zap.Int("column index", j),
						zap.String("value", cell.Value),
						zap.Any("type", cell.Type()),
						zap.String("column", header.Name))
				}
			}
			if i == 0 {
				c.logger.Info("headers of sheet",
					zap.String("sheet", sheet.Name))
				continue
			}
			rows = append(rows, rowData)
		}
		c.logger.Info("query is ready")

		var columns []string
		for _, header := range headers {
			columns = append(columns, header.Name)
		}

		c.logger.Info("columns", zap.Strings("columns", columns))
		output, err := template.Execute(pongo2.Context{
			"Table":   shared.NormalizeKey(sheet.Name),
			"Columns": columns,
			"Rows":    rows,
		})
		if err != nil {
			c.logger.Error("error parsing template", zap.Error(err))
			return 1
		}
		c.logger.Info("output generated", zap.String("query", output))

		f, err := os.Create(shared.NormalizeKey(sheet.Name) + ".sql")
		if err != nil {
			c.logger.Error("error generating file", zap.Error(err))
			return 1
		}

		w := bufio.NewWriter(f)
		if _, err := w.WriteString(output); err != nil {
			c.logger.Error("error writing file", zap.Error(err))
			return 1
		}

		_ = w.Flush()
	}

	return 0
}
