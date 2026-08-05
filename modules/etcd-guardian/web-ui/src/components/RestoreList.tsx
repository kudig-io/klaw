import { useState } from 'react'

const RestoreList = () => {
  // Mock data for restores
  const [restores, setRestores] = useState([
    { 
      id: 1, 
      name: 'restore-20260210', 
      backupName: 'daily-backup-20260209', 
      status: 'Completed', 
      time: '2026-02-10 08:30:00',
      etcdCluster: 'https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379'
    },
    { 
      id: 2, 
      name: 'restore-20260209', 
      backupName: 'daily-backup-20260208', 
      status: 'Completed', 
      time: '2026-02-09 14:00:00',
      etcdCluster: 'https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379'
    },
    { 
      id: 3, 
      name: 'restore-20260208', 
      backupName: 'daily-backup-20260207', 
      status: 'Failed', 
      time: '2026-02-08 10:15:00',
      etcdCluster: 'https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379'
    },
  ])

  const [filter, setFilter] = useState('all')

  const filteredRestores = filter === 'all' 
    ? restores 
    : restores.filter(restore => restore.status === filter)

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-blue-400">Restores</h2>
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
              <th className="py-2 px-4">Backup Name</th>
              <th className="py-2 px-4">Status</th>
              <th className="py-2 px-4">Time</th>
              <th className="py-2 px-4">Etcd Cluster</th>
              <th className="py-2 px-4">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredRestores.map((restore) => (
              <tr key={restore.id} className="border-b border-gray-600">
                <td className="py-2 px-4">{restore.name}</td>
                <td className="py-2 px-4">{restore.backupName}</td>
                <td className={`py-2 px-4 ${restore.status === 'Completed' ? 'text-green-400' : 'text-red-400'}`}>
                  {restore.status}
                </td>
                <td className="py-2 px-4">{restore.time}</td>
                <td className="py-2 px-4">{restore.etcdCluster}</td>
                <td className="py-2 px-4">
                  <div className="flex space-x-2">
                    <button className="text-blue-400 hover:text-blue-300">
                      Details
                    </button>
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

export default RestoreList