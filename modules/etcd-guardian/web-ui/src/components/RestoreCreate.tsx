import { useState } from 'react'

const RestoreCreate = () => {
  const [formData, setFormData] = useState({
    name: '',
    backupName: '',
    etcdEndpoints: '',
    etcdDataDir: '/var/lib/etcd',
    quiesceCluster: true,
  })

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Mock backup options
  const backupOptions = [
    'daily-backup-20260210',
    'daily-backup-20260209',
    'daily-backup-20260208',
    'daily-backup-20260207',
  ]

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value, type, checked } = e.target as any
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
  }

  const validateForm = () => {
    const newErrors: Record<string, string> = {}
    
    if (!formData.name) {
      newErrors.name = 'Name is required'
    }
    
    if (!formData.backupName) {
      newErrors.backupName = 'Backup name is required'
    }
    
    if (!formData.etcdEndpoints) {
      newErrors.etcdEndpoints = 'Etcd endpoints are required'
    }
    
    if (!formData.etcdDataDir) {
      newErrors.etcdDataDir = 'Etcd data directory is required'
    }
    
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) {
      return
    }
    
    setIsSubmitting(true)
    
    try {
      // Mock API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      console.log('Restore created:', formData)
      // Reset form
      setFormData({
        name: '',
        backupName: '',
        etcdEndpoints: '',
        etcdDataDir: '/var/lib/etcd',
        quiesceCluster: true,
      })
      alert('Restore created successfully!')
    } catch (error) {
      console.error('Error creating restore:', error)
      alert('Failed to create restore')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-blue-400">Create Restore</h2>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Basic Information */}
          <div className="bg-gray-700 p-4 rounded-lg shadow-md">
            <h3 className="text-lg font-semibold text-gray-300 mb-4">Basic Information</h3>
            
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Name</label>
                <input
                  type="text"
                  name="name"
                  value={formData.name}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="restore-name"
                />
                {errors.name && <p className="text-red-400 text-sm mt-1">{errors.name}</p>}
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Backup Name</label>
                <select
                  name="backupName"
                  value={formData.backupName}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select backup</option>
                  {backupOptions.map((backup, index) => (
                    <option key={index} value={backup}>{backup}</option>
                  ))}
                </select>
                {errors.backupName && <p className="text-red-400 text-sm mt-1">{errors.backupName}</p>}
              </div>
            </div>
          </div>

          {/* Etcd Configuration */}
          <div className="bg-gray-700 p-4 rounded-lg shadow-md">
            <h3 className="text-lg font-semibold text-gray-300 mb-4">Etcd Configuration</h3>
            
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Etcd Endpoints</label>
                <textarea
                  name="etcdEndpoints"
                  value={formData.etcdEndpoints}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379"
                  rows={2}
                />
                {errors.etcdEndpoints && <p className="text-red-400 text-sm mt-1">{errors.etcdEndpoints}</p>}
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Etcd Data Directory</label>
                <input
                  type="text"
                  name="etcdDataDir"
                  value={formData.etcdDataDir}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="/var/lib/etcd"
                />
                {errors.etcdDataDir && <p className="text-red-400 text-sm mt-1">{errors.etcdDataDir}</p>}
              </div>
              
              <div className="flex items-center">
                <input
                  type="checkbox"
                  name="quiesceCluster"
                  checked={formData.quiesceCluster}
                  onChange={handleChange}
                  className="w-4 h-4 text-blue-600 bg-gray-800 border-gray-600 rounded focus:ring-blue-500"
                />
                <label className="ml-2 block text-sm font-medium text-gray-300">Quiesce Cluster</label>
              </div>
            </div>
          </div>
        </div>

        {/* Submit Button */}
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={isSubmitting}
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? 'Creating...' : 'Create Restore'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default RestoreCreate