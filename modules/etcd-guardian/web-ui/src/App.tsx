import { useState } from 'react'
import { Layout, BackupList, BackupCreate, RestoreList, RestoreCreate, ScheduleList, ScheduleCreate, Dashboard } from './components'

function App() {
  const [activeTab, setActiveTab] = useState('dashboard')

  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard':
        return <Dashboard />
      case 'backups':
        return <BackupList />
      case 'create-backup':
        return <BackupCreate />
      case 'restores':
        return <RestoreList />
      case 'create-restore':
        return <RestoreCreate />
      case 'schedules':
        return <ScheduleList />
      case 'create-schedule':
        return <ScheduleCreate />
      default:
        return <Dashboard />
    }
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      {/* Header */}
      <header className="bg-gray-800 py-4 px-6 shadow-md">
        <div className="container mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-bold text-blue-400">Etcd Guardian</h1>
          <div className="flex space-x-4">
            <button 
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-md transition-colors"
              onClick={() => setActiveTab('create-backup')}
            >
              Create Backup
            </button>
            <button 
              className="px-4 py-2 bg-green-600 hover:bg-green-700 rounded-md transition-colors"
              onClick={() => setActiveTab('create-schedule')}
            >
              Create Schedule
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="container mx-auto py-6 px-4">
        {/* Navigation Tabs */}
        <div className="flex space-x-2 mb-6 overflow-x-auto pb-2">
          {[
            { id: 'dashboard', label: 'Dashboard' },
            { id: 'backups', label: 'Backups' },
            { id: 'restores', label: 'Restores' },
            { id: 'schedules', label: 'Schedules' },
          ].map((tab) => (
            <button
              key={tab.id}
              className={`px-4 py-2 rounded-md whitespace-nowrap transition-colors ${activeTab === tab.id ? 'bg-blue-600 text-white' : 'bg-gray-800 hover:bg-gray-700'}`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Content Area */}
        <div className="bg-gray-800 rounded-lg p-6 shadow-lg">
          {renderContent()}
        </div>
      </div>

      {/* Footer */}
      <footer className="bg-gray-800 py-4 px-6 mt-auto">
        <div className="container mx-auto text-center text-gray-400">
          <p>Etcd Guardian Web UI &copy; 2026</p>
        </div>
      </footer>
    </div>
  )
}

export default App