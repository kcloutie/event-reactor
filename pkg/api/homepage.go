package api

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kcloutie/event-reactor/pkg/params/version"
)

var defaultPageTemplate string = `
<!DOCTYPE html>
<html>

<head>
  <style>
    body {
      box-sizing: border-box;
      min-width: 200px;
      max-width: 800px;
      margin: 0 auto;
      padding: 45px;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji";
    }

    h1,
    h2,
    h3,
    h4,
    h5,
    h6 {
      margin-top: 24px;
      margin-bottom: 16px;
      font-weight: 600;
      line-height: 1.25;
    }

    h1 {
      margin: .67em 0;
      font-weight: 600;
      padding-bottom: .3em;
      font-size: 2em;
      border-bottom: 1px solid hsla(210, 18%, 87%, 1);
    }

    h2 {
      font-weight: 600;
      padding-bottom: .3em;
      font-size: 1.5em;
      border-bottom: 1px solid hsla(210, 18%, 87%, 1);
    }

    table {
      border: 1px solid #1C6EA4;
      background-color: #EEEEEE;

      text-align: left;
      border-collapse: collapse;
    }

    table td,
    th {
      border: 1px solid #AAAAAA;
      padding: 3px 2px;
      font-size: 18px;
    }

    table tbody td {
      font-size: 16px;
      color: #333333;
    }

    table tr:nth-child(2n) {
      background-color: #f6f8fa;
    }
  </style>
</head>
<body>
  <h1>event-reactor API Server {{VERSION}}</h1>
	<p><b>Commit: </b> {{COMMIT}}, <b>Build Date: </b> {{BUILD_DATE}}</p>
</body>
</html>
`

// AddAccount godoc
// @Summary      API Health
// @Description  API Health response
// @Tags         health
// @Accept       json
// @Produce      json
// @Param        user  body      model.HealthResponse  true  "Health"
// @Success      200      {object}  model.HealthResponse
// @Router       /health [post]
func Home(ctx context.Context, c *gin.Context) {

	buildVer := version.BuildVersion
	if buildVer == "" {
		buildVer = "nightly"
	}
	commit := version.Commit
	if commit == "" {
		commit = "unknown"
	}

	buildDate := version.BuildTime
	if buildDate == "" {
		buildDate = "unknown"
	}

	homePage := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(defaultPageTemplate, "{{VERSION}}", buildVer), "{{COMMIT}}", commit), "{{BUILD_DATE}}", buildDate)
	c.Data(200, "text/html", []byte(homePage))

}
