import { useState } from 'react'

const BackupList = () => {
  // Mock data for backups
  const [backups, setBackups] = useState([
    { 
      id: 1, 
      name: 'daily-backup-20260210', 
      mode: 'Full', 
      status: 'Completed', 
      size: '1.2 GB', 
      time: '2026-02-10 14:30:00',
      etcdRevision: 123456,
      storageLocation: 's3://my-backups/daily-backup-20260210.db',
      validation: 'Passed'
    },
    { 
      id: 2, 
      name: 'hourly-backup-20260210-13', 
      mode: 'Incremental', 
      status: 'Completed', 
      size: '24 MB', 
      time: '2026-02-10 13:00:00',
      etcdRevision: 123400,
      storageLocation: 's3://my-backups/hourly-backup-20260210-13.db',
      validation: 'Passed'
    },
    { 
      id: 3, 
      name: 'hourly-backup-20260210-12', 
      mode: 'Incremental', 
      status: 'Completed', 
      size: '18 MB', 
      time: '2026-02-10 12:00:00',
      etcdRevision: 123350,
      storageLocation: 's3://my-backups/hourly-backup-20260210-12.db',
      validation: 'Passed'
    },
    { 
      id: 4, 
      name: 'hourly-backup-20260210-11', 
      mode: 'Incremental', 
      status: 'Failed', 
      size: '0 MB', 
      time: '2026-02-10 11:00:00',
      etcdRevision: 0,
      storageLocation: '',
      validation: 'Failed'
    },
    { 
      id: 5, 
      name: 'daily-backup-20260209', 
      mode: 'Full', 
      status: 'Completed', 
      size: '1.1 GB', 
      time: '2026-02-09 00:00:00',
      etcdRevision: 122000,
      storageLocation: 's3://my-backups/daily-backup-20260209.db',
      validation: 'Passed'
    },
  ])

  const [filter, setFilter] = useState('all')

  const filteredBackups = filter === 'all' 
    ? backups 
    : backups.filter(backup => backup.status === filter)

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-blue-400">Backups</h2>
        <div className="flex space-x-2">
          <button 
            className={`px-3 py-1 rounded-md ${filter === 'all' ? 'bg-blue-600' : 'bg-gray-700'}`}
            onClick={() => setFilter('all')}
          >
            All
          </button>
          <button 
            className={`px-3 py-1 rounded-md ${filter === 'Completed' ? 'bg-green-600' : 'bg-gray-700'}`}
            onClick={() => setFilter('Completed')}
          >
            Completed
          </button>
          <button 
            className={`px-3 py-1 rounded-md ${filter === 'Failed' ? 'bg-red-600' : 'bg-gray-700'}`}
            onClick={() => setFilter('Failed')}
          >
            Failed
          </button>
        </div>
      </div>

      <div className="bg-gray-700 p-4 rounded-lg shadow-md overflow-x-auto">
        <table className="w-full text-left">
          <thead>
            <tr className="border-b border-gray-600">
              <th className="py-2 px-4">Name</th>
              <th className="py-2 px-4">Mode</th>
              <th className="py-2 px-4">Status</th>
              <th className="py-2 px-4">Size</th>
              <th className="py-2 px-4">Time</th>
              <th className="py-2 px-4">Etcd Revision</th>
              <th className="py-2 px-4">Validation</th>
              <th className="py-2 px-4">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredBackups.map((backup) => (
              <tr key={backup.id} className="border-b border-gray-600">
                <td className="py-2 px-4">{backup.name}</td>
                <td className="py-2 px-4">{backup.mode}</td>
                <td className={`py-2 px-4 ${backup.status === 'Completed' ? 'text-green-400' : 'text-red-400'}`}>
                  {backup.status}
                </td>
                <td className="py-2 px-4">{backup.size}</td>
                <td className="py-2 px-4">{backup.time}</td>
                <td className="py-2 px-4">{backup.etcdRevision}</td>
                <td className={`py-2 px-4 ${backup.validation === 'Passed' ? 'text-green-400' : 'text-red-400'}`}>
                  {backup.validation}
                </td>
                <td className="py-2 px-4">
                  <div className="flex space-x-2">
                    <button className="text-blue-400 hover:text-blue-300">
                      Details
                    </button>
                    {backup.status === 'Completed' && (
                      <button className="text-green-400 hover:text-green-300">
                        Restore
                      </button>
                    )}
                    <button className="text-red-400 hover:text-red-300">
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default BackupList