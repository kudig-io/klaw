import { useState } from 'react'

const BackupCreate = () => {
  const [formData, setFormData] = useState({
    name: '',
    mode: 'Full',
    storageProvider: 'S3',
    bucket: '',
    region: '',
    credentialsSecret: '',
    etcdEndpoints: '',
    validation: true,
    consistencyCheck: true,
  })

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

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
    
    if (!formData.bucket) {
      newErrors.bucket = 'Bucket is required'
    }
    
    if (!formData.region) {
      newErrors.region = 'Region is required'
    }
    
    if (!formData.credentialsSecret) {
      newErrors.credentialsSecret = 'Credentials secret is required'
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
      console.log('Backup created:', formData)
      // Reset form
      setFormData({
        name: '',
        mode: 'Full',
        storageProvider: 'S3',
        bucket: '',
        region: '',
        credentialsSecret: '',
        etcdEndpoints: '',
        validation: true,
        consistencyCheck: true,
      })
      alert('Backup created successfully!')
    } catch (error) {
      console.error('Error creating backup:', error)
      alert('Failed to create backup')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-blue-400">Create Backup</h2>
      
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
                  placeholder="backup-name"
                />
                {errors.name && <p className="text-red-400 text-sm mt-1">{errors.name}</p>}
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Backup Mode</label>
                <select
                  name="mode"
                  value={formData.mode}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="Full">Full</option>
                  <option value="Incremental">Incremental</option>
                </select>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Etcd Endpoints (optional)</label>
                <textarea
                  name="etcdEndpoints"
                  value={formData.etcdEndpoints}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="https://etcd-0:2379,https://etcd-1:2379"
                  rows={2}
                />
              </div>
            </div>
          </div>

          {/* Storage Configuration */}
          <div className="bg-gray-700 p-4 rounded-lg shadow-md">
            <h3 className="text-lg font-semibold text-gray-300 mb-4">Storage Configuration</h3>
            
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Storage Provider</label>
                <select
                  name="storageProvider"
                  value={formData.storageProvider}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="S3">S3</option>
                  <option value="OSS">OSS</option>
                  <option value="GCS">GCS</option>
                  <option value="Azure">Azure</option>
                </select>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Bucket</label>
                <input
                  type="text"
                  name="bucket"
                  value={formData.bucket}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="my-backups"
                />
                {errors.bucket && <p className="text-red-400 text-sm mt-1">{errors.bucket}</p>}
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Region</label>
                <input
                  type="text"
                  name="region"
                  value={formData.region}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="us-east-1"
                />
                {errors.region && <p className="text-red-400 text-sm mt-1">{errors.region}</p>}
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Credentials Secret</label>
                <input
                  type="text"
                  name="credentialsSecret"
                  value={formData.credentialsSecret}
                  onChange={handleChange}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="s3-credentials"
                />
                {errors.credentialsSecret && <p className="text-red-400 text-sm mt-1">{errors.credentialsSecret}</p>}
              </div>
            </div>
          </div>
        </div>

        {/* Validation Settings */}
        <div className="bg-gray-700 p-4 rounded-lg shadow-md">
          <h3 className="text-lg font-semibold text-gray-300 mb-4">Validation Settings</h3>
          
          <div className="space-y-3">
            <div className="flex items-center">
              <input
                type="checkbox"
                name="validation"
                checked={formData.validation}
                onChange={handleChange}
                className="w-4 h-4 text-blue-600 bg-gray-800 border-gray-600 rounded focus:ring-blue-500"
              />
              <label className="ml-2 block text-sm font-medium text-gray-300">Enable Validation</label>
            </div>
            
            {formData.validation && (
              <div className="flex items-center">
                <input
                  type="checkbox"
                  name="consistencyCheck"
                  checked={formData.consistencyCheck}
                  onChange={handleChange}
                  className="w-4 h-4 text-blue-600 bg-gray-800 border-gray-600 rounded focus:ring-blue-500"
                />
                <label className="ml-2 block text-sm font-medium text-gray-300">Enable Consistency Check</label>
              </div>
            )}
          </div>
        </div>

        {/* Submit Button */}
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={isSubmitting}
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? 'Creating...' : 'Create Backup'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default BackupCreate