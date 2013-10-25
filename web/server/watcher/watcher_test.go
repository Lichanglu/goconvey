package watcher

import (
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/smartystreets/goconvey/web/server/contract"
	"github.com/smartystreets/goconvey/web/server/system"
)

func TestWatcher(t *testing.T) {
	var (
		fixture         *watcherFixture
		expectedWatches interface{}
		actualWatches   interface{}
		expectedError   interface{}
		actualError     interface{}
	)

	Convey("Subject: Watcher", t, func() {
		fixture = newWatcherFixture()

		Convey("When initialized there should be ZERO watched folders", func() {
			So(len(fixture.watched()), ShouldEqual, 0)
		})

		Convey("When pointing to a folder", func() {
			actualWatches, expectedWatches = fixture.pointToExistingRoot("/root")

			Convey("That folder should be included as the first watched folder", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When pointing to a folder that does not exist", func() {
			actualError, expectedError = fixture.pointToImaginaryRoot("/not/there")

			Convey("An appropriate error should be returned", func() {
				So(actualError, ShouldResemble, expectedError)
			})
		})

		Convey("When pointing to a folder with nested folders", func() {
			actualWatches, expectedWatches = fixture.pointToExistingRootWithNestedFolders()

			Convey("All nested folders should be added recursively to the watched folders", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When the watcher is notified of a newly created folder", func() {
			actualWatches, expectedWatches = fixture.receiveNotificationOfNewFolder()

			Convey("The folder should be included in the watched folders", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When the watcher is notified of a recently deleted folder", func() {
			actualWatches, expectedWatches = fixture.receiveNotificationOfDeletedFolder()

			Convey("The folder should no longer be included in the watched folders", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When a watched folder is ignored", func() {
			actualWatches, expectedWatches = fixture.ignoreWatchedFolder()

			Convey("The folder should not be included in the watched folders", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When a folder that is not being watched is ignored", func() {
			actualWatches, expectedWatches = fixture.ignoreIrrelevantFolder()

			Convey("The request should be ignored", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When a folder that does not exist is ignored", func() {
			actualWatches, expectedWatches = fixture.ignoreImaginaryFolder()

			Convey("There should be no change to the watched folders", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When an ignored folder is reinstated", func() {
			actualWatches, expectedWatches = fixture.reinstateIgnoredFolder()

			Convey("The folder should be included once more in the watched folders", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When an ignored folder is deleted and then reinstated", func() {
			actualWatches, expectedWatches = fixture.reinstateDeletedFolder()

			Convey("The reinstatement request should be ignored", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("When a folder that is not being watched is reinstated", func() {
			actualWatches, expectedWatches = fixture.reinstateIrrelevantFolder()

			Convey("The request should be ignored", func() {
				So(actualWatches, ShouldResemble, expectedWatches)
			})
		})

		Convey("Regardless of the status of the watched folders", func() {
			folders := fixture.setupSeveralFoldersWithWatcher()

			Convey("The IsActive query method should conform to the actual state returned", func() {
				So(fixture.watcher.IsActive(folders["active"]), ShouldBeTrue)
				So(fixture.watcher.IsActive(folders["reinstated"]), ShouldBeTrue)

				So(fixture.watcher.IsActive(folders["ignored"]), ShouldBeFalse)
				So(fixture.watcher.IsActive(folders["deleted"]), ShouldBeFalse)
				So(fixture.watcher.IsActive(folders["irrelevant"]), ShouldBeFalse)
			})

			Convey("The IsIgnored query method should conform to the actual state returned", func() {
				So(fixture.watcher.IsIgnored(folders["ignored"]), ShouldBeTrue)

				So(fixture.watcher.IsIgnored(folders["active"]), ShouldBeFalse)
				So(fixture.watcher.IsIgnored(folders["reinstated"]), ShouldBeFalse)
				So(fixture.watcher.IsIgnored(folders["deleted"]), ShouldBeFalse)
				So(fixture.watcher.IsIgnored(folders["irrelevant"]), ShouldBeFalse)
			})

			Convey("The IsWatched query method should conform to the actual state returned", func() {
				So(fixture.watcher.IsWatched(folders["active"]), ShouldBeTrue)
				So(fixture.watcher.IsWatched(folders["reinstated"]), ShouldBeTrue)
				So(fixture.watcher.IsWatched(folders["ignored"]), ShouldBeTrue)

				So(fixture.watcher.IsWatched(folders["deleted"]), ShouldBeFalse)
				So(fixture.watcher.IsWatched(folders["irrelevant"]), ShouldBeFalse)
			})
		})
	})
}

type watcherFixture struct {
	watcher *Watcher
	fs      *system.FakeFileSystem
}

func (self *watcherFixture) watched() []*contract.Package {
	return self.watcher.WatchedFolders()
}

func (self *watcherFixture) verifyQueryMethodsInSync() bool {
	return false
}

func (self *watcherFixture) pointToExistingRoot(folder string) (actual, expected interface{}) {
	self.fs.Create(folder, 1, time.Now())

	self.watcher.Adjust(folder)

	actual = self.watched()
	expected = []*contract.Package{&contract.Package{Active: true, Path: "/root", Name: "root"}}
	return
}

func (self *watcherFixture) pointToImaginaryRoot(folder string) (actual, expected interface{}) {
	actual = self.watcher.Adjust(folder)
	expected = errors.New("Directory does not exist: '/not/there'")
	return
}

func (self *watcherFixture) pointToExistingRootWithNestedFolders() (actual, expected interface{}) {
	self.fs.Create("/root", 1, time.Now())
	self.fs.Create("/root/sub", 2, time.Now())
	self.fs.Create("/root/sub2", 3, time.Now())
	self.fs.Create("/root/sub/subsub", 4, time.Now())

	self.watcher.Adjust("/root")

	actual = self.watched()
	expected = []*contract.Package{
		&contract.Package{Active: true, Path: "/root", Name: "root"},
		&contract.Package{Active: true, Path: "/root/sub", Name: "sub"},
		&contract.Package{Active: true, Path: "/root/sub2", Name: "sub2"},
		&contract.Package{Active: true, Path: "/root/sub/subsub", Name: "subsub"},
	}
	return
}

func (self *watcherFixture) receiveNotificationOfNewFolder() (actual, expected interface{}) {
	self.watcher.Creation("/root/sub")

	actual = self.watched()
	expected = []*contract.Package{&contract.Package{Active: true, Path: "/root/sub", Name: "sub"}}
	return
}

func (self *watcherFixture) receiveNotificationOfDeletedFolder() (actual, expected interface{}) {
	self.watcher.Creation("/root/sub2")
	self.watcher.Creation("/root/sub")

	self.watcher.Deletion("/root/sub")

	actual = self.watched()
	expected = []*contract.Package{&contract.Package{Active: true, Path: "/root/sub2", Name: "sub2"}}
	return
}

func (self *watcherFixture) ignoreWatchedFolder() (actual, expected interface{}) {
	self.watcher.Creation("/root/sub2")

	self.watcher.Ignore("/root/sub2")

	actual = self.watched()
	expected = []*contract.Package{&contract.Package{Active: false, Path: "/root/sub2", Name: "sub2"}}
	return
}

func (self *watcherFixture) ignoreIrrelevantFolder() (actual, expected interface{}) {
	self.fs.Create("/root", 1, time.Now())
	self.fs.Create("/something", 1, time.Now())
	self.watcher.Adjust("/root")

	self.watcher.Ignore("/something")

	actual = self.watched()
	expected = []*contract.Package{&contract.Package{Active: true, Path: "/root", Name: "root"}}
	return
}

func (self *watcherFixture) ignoreImaginaryFolder() (actual, expected interface{}) {
	self.fs.Create("/root", 1, time.Now())
	self.watcher.Adjust("/root")

	self.watcher.Ignore("/not/there")

	actual = self.watched()
	expected = []*contract.Package{&contract.Package{Active: true, Path: "/root", Name: "root"}}
	return
}

func (self *watcherFixture) reinstateIgnoredFolder() (actual, expected interface{}) {
	self.fs.Create("/root", 1, time.Now())
	self.fs.Create("/root/sub", 2, time.Now())
	self.watcher.Adjust("/root")
	self.watcher.Ignore("/root/sub")

	self.watcher.Reinstate("/root/sub")

	actual = self.watched()
	expected = []*contract.Package{
		&contract.Package{Active: true, Path: "/root", Name: "root"},
		&contract.Package{Active: true, Path: "/root/sub", Name: "sub"},
	}
	return
}

func (self *watcherFixture) reinstateDeletedFolder() (actual, expected interface{}) {
	self.fs.Create("/root", 1, time.Now())
	self.fs.Create("/root/sub", 2, time.Now())
	self.watcher.Adjust("/root")
	self.watcher.Ignore("/root/sub")
	self.watcher.Deletion("/root/sub")

	self.watcher.Reinstate("/root/sub")

	actual = self.watched()
	expected = []*contract.Package{&contract.Package{Active: true, Path: "/root", Name: "root"}}
	return
}

func (self *watcherFixture) reinstateIrrelevantFolder() (actual, expected interface{}) {
	self.fs.Create("/root", 1, time.Now())
	self.fs.Create("/irrelevant", 2, time.Now())
	self.watcher.Adjust("/root")

	self.watcher.Reinstate("/irrelevant")

	actual = self.watched()
	expected = []*contract.Package{&contract.Package{Active: true, Path: "/root", Name: "root"}}
	return
}

func (self *watcherFixture) setupSeveralFoldersWithWatcher() map[string]string {
	self.fs.Create("/folder", 0, time.Now())
	self.fs.Create("/folder/active", 1, time.Now())
	self.fs.Create("/folder/reinstated", 2, time.Now())
	self.fs.Create("/folder/ignored", 3, time.Now())
	self.fs.Create("/folder/deleted", 4, time.Now())
	self.fs.Create("/irrelevant", 5, time.Now())

	self.watcher.Adjust("/folder")
	self.watcher.Ignore("/folder/ignored")
	self.watcher.Ignore("/folder/reinstated")
	self.watcher.Reinstate("/folder/reinstated")
	self.watcher.Deletion("/folder/deleted")
	self.fs.Delete("/folder/deleted")

	return map[string]string{
		"active":     "/folder/active",
		"reinstated": "/folder/reinstated",
		"ignored":    "/folder/ignored",
		"deleted":    "/folder/deleted",
		"irrelevant": "/irrelevant",
	}
}

func newWatcherFixture() *watcherFixture {
	self := &watcherFixture{}
	self.fs = system.NewFakeFileSystem()
	self.watcher = NewWatcher(self.fs)
	return self
}

func init() {
	fmt.Sprintf("Keeps fmt in the import list...")
}
