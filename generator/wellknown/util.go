package wellknown

import (
	"strings"

	v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/protobuf/compiler/protogen"
)

func trimComment(comment string) string {
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(comment), "/"))
}

func handleTitleAndDescription(schema *v3.Schema, comments *protogen.CommentSet) {
	if schema == nil {
		return
	}

	if comments == nil {
		return
	}

	if comments.Leading != "" {
		schema.Title = trimComment(comments.Leading.String())
	}

	schema.Description += trimComment(comments.Leading.String()) + "\n"
	for _, v := range comments.LeadingDetached {
		schema.Description += trimComment(v.String()) + "\n"
	}

	if comments.Trailing.String() != "" {
		schema.Description = trimComment(comments.Trailing.String())
	}
}
