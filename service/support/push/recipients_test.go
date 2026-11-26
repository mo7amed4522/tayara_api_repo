package push_test

import (
	"testing"

	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	"github.com/mo7amed4522/tayara_api/service/support/push"
)

func strptr(v uint) *uint { return &v }

func TestResolveRecipientsAssignedEmployee(t *testing.T) {
	ticket := supportModel.Ticket{CustomerID: 51, AssignedEmployeeID: strptr(9), AssignedAuthorityID: strptr(4)}
	got := push.ResolveRecipients("customer", 51, ticket, []push.Participant{{ActorType: "support", ActorID: 77}})
	if len(got) != 1 || got[0].ActorType != "support" || got[0].ActorID != 9 {
		t.Fatalf("assigned employee must be the only recipient: %+v", got)
	}
}

func TestResolveRecipientsUnassignedQueue(t *testing.T) {
	ticket := supportModel.Ticket{CustomerID: 51}
	queue := []push.Participant{{ActorType: "support", ActorID: 9}, {ActorType: "quality", ActorID: 3}}
	got := push.ResolveRecipients("customer", 51, ticket, queue)
	if len(got) != 2 {
		t.Fatalf("unassigned ticket must fan out to the queue: %+v", got)
	}
}

func TestResolveRecipientsStaffSenderTargetsCustomer(t *testing.T) {
	ticket := supportModel.Ticket{CustomerID: 51, AssignedEmployeeID: strptr(9)}
	got := push.ResolveRecipients("support", 9, ticket, nil)
	if len(got) != 1 || got[0].ActorType != "customer" || got[0].ActorID != 51 {
		t.Fatalf("staff message must target the customer: %+v", got)
	}
}

func TestResolveRecipientsNeverSender(t *testing.T) {
	ticket := supportModel.Ticket{CustomerID: 51}
	queue := []push.Participant{{ActorType: "support", ActorID: 9}}
	got := push.ResolveRecipients("support", 9, ticket, queue)
	for _, p := range got {
		if p.ActorType == "support" && p.ActorID == 9 {
			t.Fatal("sender must never receive its own push")
		}
	}
}
