import { useState } from 'react'

const ScheduleList = () => {
  // Mock data for schedules
  const [schedules, setSchedules] = useState([
    { 
      id: 1, 
      name: 'daily-backup', 
      schedule: '0 0 * * *', 
      mode: 'Full', 
      lastRun: '2026-02-10 00:00:00',
      nextRun: '2026-02-11 00:00:00',
      status: 'Active'
    },
    { 
      id: 2, 
      name: 'hourly-backup', 
      schedule: '0 * * * *', 
      mode: 'Incremental', 
      lastRun: '2026-02-10 14:00:00',
      nextRun: '2026-02-10 15:00:00',
      status: 'Active'
    },
    { 
      id: 3, 
      name: 'weekly-backup', 
      schedule: '0 0 * * 0', 
      mode: 'Full', 
      lastRun: '2026-02-08 00:00:00',
      nextRun: '2026-02-15 00:00:00',
      status: 'Active'
    },
  ])

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-blue-400">Schedules</h2>

      <div className="bg-gray-700 p-4 rounded-lg shadow-md overflow-x-auto">
        <table className="w-full text-left">
          <thead>
            <tr className="border-b border-gray-600">
              <th className="py-2 px-4">Name</th>
              <th className="py-2 px-4">Schedule</th>
              <th className="py-2 px-4">Mode</th>
              <th className="py-2 px-4">Last Run</th>
              <th className="py-2 px-4">Next Run</th>
              <th className="py-2 px-4">Status</th>
              <th className="py-2 px-4">Actions</th>
            </tr>
          </thead>
          <tbody>
            {schedules.map((schedule) => (
              <tr key={schedule.id} className="border-b border-gray-600">
                <td className="py-2 px-4">{schedule.name}</td>
                <td className="py-2 px-4">{schedule.schedule}</td>
                <td className="py-2 px-4">{schedule.mode}</td>
                <td className="py-2 px-4">{schedule.lastRun}</td>
                <td className="py-2 px-4">{schedule.nextRun}</td>
                <td className="py-2 px-4 text-green-400">{schedule.status}</td>
                <td className="py-2 px-4">
                  <div className="flex space-x-2">
                    <button className="text-blue-400 hover:text-blue-300">
                      Edit
                    </button>
                    <button className="text-green-400 hover:text-green-300">
                      Run Now
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

export default ScheduleList