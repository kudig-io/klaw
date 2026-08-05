import { useState } from 'react'

const Dashboard = () => {
  // Mock data for dashboard
  const [metrics, setMetrics] = useState({
    totalBackups: 42,
    successfulBackups: 38,
    failedBackups: 4,
    totalRestores: 8,
    successfulRestores: 7,
    failedRestores: 1,
    etcdSize: '2.4 GB',
    lastBackupTime: '2026-02-10 14:30:00',
  })

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-blue-400">Dashboard</h2>
      
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div className="bg-gray-700 p-4 rounded-lg shadow-md">
          <h3 className="text-lg font-semibold text-gray-300">Total Backups</h3>
          <p className="text-3xl font-bold text-white">{metrics.totalBackups}</p>
          <div className="flex space-x-2 mt-2">
            <span className="text-green-400 text-sm">✓ {metrics.successfulBackups} Successful</span>
            <span className="text-red-400 text-sm">✗ {metrics.failedBackups} Failed</span>
          </div>
        </div>

        <div className="bg-gray-700 p-4 rounded-lg shadow-md">
          <h3 className="text-lg font-semibold text-gray-300">Total Restores</h3>
          <p className="text-3xl font-bold text-white">{metrics.totalRestores}</p>
          <div className="flex space-x-2 mt-2">
            <span className="text-green-400 text-sm">✓ {metrics.successfulRestores} Successful</span>
            <span className="text-red-400 text-sm">✗ {metrics.failedRestores} Failed</span>
          </div>
        </div>

        <div className="bg-gray-700 p-4 rounded-lg shadow-md">
          <h3 className="text-lg font-semibold text-gray-300">Etcd Size</h3>
          <p className="text-3xl font-bold text-white">{metrics.etcdSize}</p>
          <p className="text-gray-400 text-sm mt-2">Last Backup: {metrics.lastBackupTime}</p>
        </div>
      </div>

      {/* Recent Backups */}
      <div className="bg-gray-700 p-4 rounded-lg shadow-md">
        <h3 className="text-lg font-semibold text-gray-300 mb-4">Recent Backups</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-gray-600">
                <th className="py-2 px-4">Name</th>
                <th className="py-2 px-4">Mode</th>
                <th className="py-2 px-4">Status</th>
                <th className="py-2 px-4">Size</th>
                <th className="py-2 px-4">Time</th>
              </tr>
            </thead>
            <tbody>
              {[
                { name: 'daily-backup-20260210', mode: 'Full', status: 'Completed', size: '1.2 GB', time: '2026-02-10 14:30:00' },
                { name: 'hourly-backup-20260210-13', mode: 'Incremental', status: 'Completed', size: '24 MB', time: '2026-02-10 13:00:00' },
                { name: 'hourly-backup-20260210-12', mode: 'Incremental', status: 'Completed', size: '18 MB', time: '2026-02-10 12:00:00' },
                { name: 'hourly-backup-20260210-11', mode: 'Incremental', status: 'Failed', size: '0 MB', time: '2026-02-10 11:00:00' },
              ].map((backup, index) => (
                <tr key={index} className="border-b border-gray-600">
                  <td className="py-2 px-4">{backup.name}</td>
                  <td className="py-2 px-4">{backup.mode}</td>
                  <td className={`py-2 px-4 ${backup.status === 'Completed' ? 'text-green-400' : 'text-red-400'}`}>
                    {backup.status}
                  </td>
                  <td className="py-2 px-4">{backup.size}</td>
                  <td className="py-2 px-4">{backup.time}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Schedules */}
      <div className="bg-gray-700 p-4 rounded-lg shadow-md">
        <h3 className="text-lg font-semibold text-gray-300 mb-4">Active Schedules</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-gray-600">
                <th className="py-2 px-4">Name</th>
                <th className="py-2 px-4">Schedule</th>
                <th className="py-2 px-4">Mode</th>
                <th className="py-2 px-4">Last Run</th>
                <th className="py-2 px-4">Status</th>
              </tr>
            </thead>
            <tbody>
              {[
                { name: 'daily-backup', schedule: '0 0 * * *', mode: 'Full', lastRun: '2026-02-10 00:00:00', status: 'Active' },
                { name: 'hourly-backup', schedule: '0 * * * *', mode: 'Incremental', lastRun: '2026-02-10 14:00:00', status: 'Active' },
              ].map((schedule, index) => (
                <tr key={index} className="border-b border-gray-600">
                  <td className="py-2 px-4">{schedule.name}</td>
                  <td className="py-2 px-4">{schedule.schedule}</td>
                  <td className="py-2 px-4">{schedule.mode}</td>
                  <td className="py-2 px-4">{schedule.lastRun}</td>
                  <td className="py-2 px-4 text-green-400">{schedule.status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

export default Dashboard