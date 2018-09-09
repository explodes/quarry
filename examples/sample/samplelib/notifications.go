package samplelib

import (
	"context"
	"fmt"

	"github.com/explodes/quarry"
	"github.com/explodes/quarry/examples/sample/samplepb"
	"github.com/explodes/quarry/examples/sample/samplequarry"
)

func init() {
	graph := samplequarry.Default()

	// Dependencies for NotificationService would normally be provided by the graph.
	// For a service-like object, consider a quarry.Singleton.
	notificationService := &NotificationService{}
	graph.MustAddFactory("notificationService", quarry.Provider(notificationService))

	graph.MustAddFactory("notifications", fetchNotifications)
	graph.MustAddDependency("notifications", "notificationService")
	graph.MustAddDependency("notifications", "user")

	graph.MustAddFactory("unreadNotifications", fetchUnreadNotifications)
	graph.MustAddDependency("unreadNotifications", "notifications")

	graph.MustAddFactory("inbox", fetchInbox)
	graph.MustAddDependency("inbox", "notifications")
	// unreadOption indicates that this dependency is only filled when the unread option
	// is passed in by parameters.
	graph.MustAddDependency("inbox", "unreadNotifications", unreadOption)

}

type NotificationService struct{}

func (s *NotificationService) FetchNotifications(ctx context.Context, user *samplepb.User) ([]*samplepb.Notification, error) {
	fmt.Println("NotificationService::FetchNotifications")
	// Simulate using context.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// "Fetch" notifications.
	admin := &samplepb.User{
		Username: "admin",
		Email:    "admin@example.com",
	}
	notifications := []*samplepb.Notification{
		{Title: "hello", Body: fmt.Sprintf("Hello, %s!", user.Username), To: user, From: admin, Status: samplepb.Notification_UNREAD},
		{Title: "graph time", Body: "Graphs are pretty fun", To: user, From: admin, Status: samplepb.Notification_READ},
	}
	return notifications, nil
}

func fetchNotifications(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
	user := deps["user"].(*samplepb.User)
	notificationService := deps["notificationService"].(*NotificationService)

	return notificationService.FetchNotifications(ctx, user)
}

func unreadOption(params interface{}) bool {
	request := params.(*samplepb.SampleRequest)
	return request.ShowUnread
}

func fetchUnreadNotifications(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
	notifications := deps["notifications"].([]*samplepb.Notification)

	var unread []*samplepb.Notification
	for _, message := range notifications {
		if message.Status != samplepb.Notification_READ {
			unread = append(unread, message)
		}
	}

	return unread, nil
}

func fetchInbox(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
	notifications := deps["notifications"].([]*samplepb.Notification)
	// unreadNotifications is conditional, so it could come in as a null value.
	unreadNotifications, _ := deps["unreadNotifications"].([]*samplepb.Notification)

	inbox := &samplepb.Inbox{
		Notifications:       notifications,
		UnreadNotifications: unreadNotifications,
	}
	return inbox, nil
}
