package action_test

import (
	"errors"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/goulash/slackapi/slackapifakes"
	"github.com/pivotalservices/slack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DisableUser", func() {
	var (
		a            action.Action
		c            config.Config
		fakeSlackAPI *slackapifakes.FakeSlackAPI
		logger       lager.Logger
	)

	BeforeEach(func() {
		fakeSlackAPI = &slackapifakes.FakeSlackAPI{}
		c = config.NewLocalConfig(
			"slack-auth-token",
			"/slack-slash-command",
			"slack-team-name",
			"slack-user-id",
			"audit-log-channel-id",
			"uninvitable-domain.com",
			"uninvitable-domain-message",
		)

		logger = lager.NewLogger("testlogger")
	})

	Describe("Do", func() {
		It("attempts to disable the user if they can be found by name", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:           "U1234",
					Name:         "tsmith",
					IsRestricted: true,
				},
			}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user @tsmith",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fakeSlackAPI.GetUsersCallCount()).Should(Equal(1))
			Ω(fakeSlackAPI.DisableUserCallCount()).Should(Equal(1))

			actualSlackTeamName, actualID := fakeSlackAPI.DisableUserArgsForCall(0)
			Ω(actualSlackTeamName).Should(Equal("slack-team-name"))
			Ω(actualID).Should(Equal("U1234"))
		})

		It("attempts to disable the user if they can be found by email", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:           "U1234",
					IsRestricted: true,
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fakeSlackAPI.GetUsersCallCount()).Should(Equal(1))
			Ω(fakeSlackAPI.DisableUserCallCount()).Should(Equal(1))

			actualSlackTeamName, actualID := fakeSlackAPI.DisableUserArgsForCall(0)
			Ω(actualSlackTeamName).Should(Equal("slack-team-name"))
			Ω(actualID).Should(Equal("U1234"))
		})

		It("returns an error if the GetUsers call fails", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{}, errors.New("error"))

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("error"))
			Ω(result).Should(Equal("Failed to disable user 'user@example.com': error"))
		})

		It("returns an error if the user cannot be found", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Unable to find user matching 'user@example.com'."))
			Ω(result).Should(Equal("Failed to disable user 'user@example.com': Unable to find user matching 'user@example.com'."))
		})

		It("returns an error when disabling the user fails", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:           "U1234",
					IsRestricted: true,
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			fakeSlackAPI.DisableUserReturns(errors.New("failed"))

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(result).Should(Equal("Failed to disable user 'user@example.com': failed"))
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("failed"))
		})

		It("returns nil on success", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:           "U1234",
					IsRestricted: true,
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Successfully disabled user 'user@example.com'"))
		})

		It("returns an error if the user is a full user", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U9999",
					IsRestricted:      false,
					IsUltraRestricted: false,
					Profile: slack.UserProfile{
						Email: "jones@example.com",
					},
				},
				{
					ID:                "U1234",
					IsRestricted:      false,
					IsUltraRestricted: false,
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(result).Should(Equal("Failed to disable user 'user@example.com': Full users cannot be disabled."))
		})

		It("does not return an error if another user is a full user", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U9999",
					IsRestricted:      false,
					IsUltraRestricted: false,
					Profile: slack.UserProfile{
						Email: "jones@example.com",
					},
				},
				{
					ID:                "U1234",
					IsRestricted:      false,
					IsUltraRestricted: true,
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("AuditMessage", func() {
		It("exists", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			aa, ok := a.(action.AuditableAction)
			Ω(ok).Should(BeTrue())

			Ω(aa.AuditMessage(fakeSlackAPI)).Should(Equal("@commander-name disabled user user@example.com"))
		})
	})
})
