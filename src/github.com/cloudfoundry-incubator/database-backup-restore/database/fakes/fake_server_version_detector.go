// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/database"
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/version"
)

type FakeServerVersionDetector struct {
	GetVersionStub        func(config.ConnectionConfig, config.TempFolderManager) (version.DatabaseServerVersion, error)
	getVersionMutex       sync.RWMutex
	getVersionArgsForCall []struct {
		arg1 config.ConnectionConfig
		arg2 config.TempFolderManager
	}
	getVersionReturns struct {
		result1 version.DatabaseServerVersion
		result2 error
	}
	getVersionReturnsOnCall map[int]struct {
		result1 version.DatabaseServerVersion
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeServerVersionDetector) GetVersion(arg1 config.ConnectionConfig, arg2 config.TempFolderManager) (version.DatabaseServerVersion, error) {
	fake.getVersionMutex.Lock()
	ret, specificReturn := fake.getVersionReturnsOnCall[len(fake.getVersionArgsForCall)]
	fake.getVersionArgsForCall = append(fake.getVersionArgsForCall, struct {
		arg1 config.ConnectionConfig
		arg2 config.TempFolderManager
	}{arg1, arg2})
	fake.recordInvocation("GetVersion", []interface{}{arg1, arg2})
	fake.getVersionMutex.Unlock()
	if fake.GetVersionStub != nil {
		return fake.GetVersionStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getVersionReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeServerVersionDetector) GetVersionCallCount() int {
	fake.getVersionMutex.RLock()
	defer fake.getVersionMutex.RUnlock()
	return len(fake.getVersionArgsForCall)
}

func (fake *FakeServerVersionDetector) GetVersionCalls(stub func(config.ConnectionConfig, config.TempFolderManager) (version.DatabaseServerVersion, error)) {
	fake.getVersionMutex.Lock()
	defer fake.getVersionMutex.Unlock()
	fake.GetVersionStub = stub
}

func (fake *FakeServerVersionDetector) GetVersionArgsForCall(i int) (config.ConnectionConfig, config.TempFolderManager) {
	fake.getVersionMutex.RLock()
	defer fake.getVersionMutex.RUnlock()
	argsForCall := fake.getVersionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServerVersionDetector) GetVersionReturns(result1 version.DatabaseServerVersion, result2 error) {
	fake.getVersionMutex.Lock()
	defer fake.getVersionMutex.Unlock()
	fake.GetVersionStub = nil
	fake.getVersionReturns = struct {
		result1 version.DatabaseServerVersion
		result2 error
	}{result1, result2}
}

func (fake *FakeServerVersionDetector) GetVersionReturnsOnCall(i int, result1 version.DatabaseServerVersion, result2 error) {
	fake.getVersionMutex.Lock()
	defer fake.getVersionMutex.Unlock()
	fake.GetVersionStub = nil
	if fake.getVersionReturnsOnCall == nil {
		fake.getVersionReturnsOnCall = make(map[int]struct {
			result1 version.DatabaseServerVersion
			result2 error
		})
	}
	fake.getVersionReturnsOnCall[i] = struct {
		result1 version.DatabaseServerVersion
		result2 error
	}{result1, result2}
}

func (fake *FakeServerVersionDetector) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getVersionMutex.RLock()
	defer fake.getVersionMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeServerVersionDetector) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ database.ServerVersionDetector = new(FakeServerVersionDetector)
