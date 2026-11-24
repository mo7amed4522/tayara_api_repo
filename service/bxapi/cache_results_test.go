package bxapi

import (
	"testing"

	"github.com/mo7amed4522/tayara_api/model/bxapi/response"
	"github.com/mo7amed4522/tayara_api/model/catalog"
)

func TestRecommendationWishlistFlagsAreOverlaidPerUser(t *testing.T) {
	publicPool := []response.CustomerRecommendedResult{
		{Package: catalog.Package{ID: 11}},
		{Package: catalog.Package{ID: 22}},
	}
	userOne := append([]response.CustomerRecommendedResult(nil), publicPool...)
	userTwo := append([]response.CustomerRecommendedResult(nil), publicPool...)

	applyWishlistFlags(userOne, []uint{11})
	applyWishlistFlags(userTwo, nil)

	if !userOne[0].Wishlisted || userOne[1].Wishlisted {
		t.Fatalf("user one flags = [%v, %v], want [true, false]", userOne[0].Wishlisted, userOne[1].Wishlisted)
	}
	if userTwo[0].Wishlisted || userTwo[1].Wishlisted {
		t.Fatalf("user two flags = [%v, %v], want [false, false]", userTwo[0].Wishlisted, userTwo[1].Wishlisted)
	}
	if publicPool[0].Wishlisted || publicPool[1].Wishlisted {
		t.Fatal("overlay mutated the shared public pool")
	}
}
