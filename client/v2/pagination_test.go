package v2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Pagination(t *testing.T) {
	t.Parallel()

	c := newTestClient(t)
	type TestType struct{}
	pagerUrl := "/2/thing/myteam/resource"

	t.Run("returns a Pager with the correct URL and options", func(t *testing.T) {
		p, err := NewPager[TestType](c, pagerUrl)
		require.NoError(t, err)

		assert.NotNil(t, p.client)
		assert.Equal(t, defaultPageSize, p.opts.PageSize)
		assert.Equal(t, fmt.Sprintf("%s?page%%5Bsize%%5D=%d", pagerUrl, defaultPageSize), *p.next)
		assert.True(t, p.HasNext())
	})

	t.Run("returns a Pager with the correct URL and options when PageSize is set", func(t *testing.T) {
		p, err := NewPager[TestType](c, pagerUrl, PageSize(50))
		require.NoError(t, err)

		assert.NotNil(t, p.client)
		assert.Equal(t, 50, p.opts.PageSize)
		assert.Equal(t, fmt.Sprintf("%s?page%%5Bsize%%5D=%d", pagerUrl, 50), *p.next)
		assert.True(t, p.HasNext())
	})
}
