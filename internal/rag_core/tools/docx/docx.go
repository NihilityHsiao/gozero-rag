package docx

import "github.com/cloudwego/eino/schema"

func IsMarkdown(doc *schema.Document) bool {
	ext, ok := doc.MetaData["_extension"]
	if !ok {
		return false
	}
	return ext == ".md" || ext == ".markdown"
}

func IsPdf(doc *schema.Document) bool {
	ext, ok := doc.MetaData["_extension"]
	if !ok {
		return false
	}
	return ext == ".pdf" || ext == ".PDF"
}
